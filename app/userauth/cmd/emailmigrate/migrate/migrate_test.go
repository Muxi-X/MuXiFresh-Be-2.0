package migrate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 连库测试：仅当设置了 MUXI_TEST_MONGO_URL 时运行，否则跳过（CI 默认不连库）。
// 本地验证：
//
//	docker run -d --rm -p 27017:27017 --name muxi-mongo-test mongo:7
//	$env:MUXI_TEST_MONGO_URL="mongodb://localhost:27017"; go test ./app/userauth/cmd/emailmigrate/...
func testClient(t *testing.T) (*mongo.Client, string) {
	t.Helper()
	url := os.Getenv("MUXI_TEST_MONGO_URL")
	if url == "" {
		t.Skip("set MUXI_TEST_MONGO_URL to run the emailmigrate integration test")
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping: %v", err)
	}
	db := fmt.Sprintf("emailmigrate_test_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_ = client.Database(db).Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	return client, db
}

var (
	oidOld  = primitive.NewObjectIDFromTimestamp(time.Date(2023, 9, 28, 0, 0, 0, 0, time.UTC))
	oidNew  = primitive.NewObjectIDFromTimestamp(time.Date(2023, 10, 8, 0, 0, 0, 0, time.UTC))
	oidSolo = primitive.NewObjectIDFromTimestamp(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
)

func seed(t *testing.T, client *mongo.Client, db string) {
	t.Helper()
	ui := client.Database(db).Collection("userinfo")
	ua := client.Database(db).Collection("userauth")

	mustInsert(t, ui,
		bson.M{"_id": oidOld, "email": "12345678@QQ.COM", "student_id": "S1"},
		bson.M{"_id": oidNew, "email": "12345678@qq.com", "student_id": ""},
		bson.M{"_id": oidSolo, "email": "Solo@EXAMPLE.com", "student_id": ""},
	)
	mustInsert(t, ua,
		bson.M{"_id": primitive.NewObjectID(), "userInfoID": oidOld, "email": "12345678@QQ.COM"},
		bson.M{"_id": primitive.NewObjectID(), "userInfoID": oidNew, "email": "12345678@qq.com"},
		bson.M{"_id": primitive.NewObjectID(), "userInfoID": oidSolo, "email": "Solo@EXAMPLE.com"},
	)
}

func mustInsert(t *testing.T, coll *mongo.Collection, docs ...bson.M) {
	t.Helper()
	for _, d := range docs {
		if _, err := coll.InsertOne(context.Background(), d); err != nil {
			t.Fatalf("insert %v: %v", d, err)
		}
	}
}

func getEmail(t *testing.T, coll *mongo.Collection, id primitive.ObjectID) (string, bool) {
	t.Helper()
	var doc bson.M
	if err := coll.FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc); err != nil {
		t.Fatalf("find %s: %v", id.Hex(), err)
	}
	v, ok := doc["email"]
	if !ok {
		return "", false
	}
	s, _ := v.(string)
	return s, true
}

func TestRun_DryRunDoesNotMutate(t *testing.T) {
	client, db := testClient(t)
	seed(t, client, db)

	var buf bytes.Buffer
	if err := Run(context.Background(), client, db, false, &buf); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if e, _ := getEmail(t, client.Database(db).Collection("userinfo"), oidOld); e != "12345678@QQ.COM" {
		t.Fatalf("dry-run must not change data, old email=%q", e)
	}
	if !bytes.Contains(buf.Bytes(), []byte("[duplicate]")) {
		t.Fatalf("dry-run should report the duplicate group, got:\n%s", buf.String())
	}
}

func TestRun_ApplyConvergesAndEnablesUniqueIndex(t *testing.T) {
	client, db := testClient(t)
	seed(t, client, db)

	ctx := context.Background()
	ui := client.Database(db).Collection("userinfo")
	ua := client.Database(db).Collection("userauth")

	var buf bytes.Buffer
	if err := Run(ctx, client, db, true, &buf); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// keeper（较新）保留邮箱，并接收旧账号学号
	if e, ok := getEmail(t, ui, oidNew); !ok || e != "12345678@qq.com" {
		t.Fatalf("keeper email wrong: %q ok=%v", e, ok)
	}
	var keeper bson.M
	if err := ui.FindOne(ctx, bson.M{"_id": oidNew}).Decode(&keeper); err != nil {
		t.Fatal(err)
	}
	if keeper["student_id"] != "S1" {
		t.Fatalf("student_id should be migrated to keeper, got %v", keeper["student_id"])
	}

	// 旧账号邮箱被摘除（字段不存在），学号已搬走
	if _, ok := getEmail(t, ui, oidOld); ok {
		t.Fatal("old userinfo email must be unset")
	}
	var oldDoc bson.M
	if err := ui.FindOne(ctx, bson.M{"_id": oidOld}).Decode(&oldDoc); err != nil {
		t.Fatal(err)
	}
	if _, ok := oldDoc["student_id"]; ok {
		t.Fatalf("old student_id must be unset after migration, got %v", oldDoc["student_id"])
	}
	var oldAuth bson.M
	if err := ua.FindOne(ctx, bson.M{"userInfoID": oidOld}).Decode(&oldAuth); err != nil {
		t.Fatal(err)
	}
	if _, ok := oldAuth["email"]; ok {
		t.Fatal("old userauth email must be unset")
	}

	// 非重复账号域名拍平
	if e, _ := getEmail(t, ui, oidSolo); e != "Solo@example.com" {
		t.Fatalf("solo email should be flattened, got %q", e)
	}

	// 治理后唯一索引可成功创建（本 PR 的核心验证点）
	for _, spec := range []mongodb.IndexSpec{
		{Collection: "userinfo", Name: "userinfo_email_unique", Unique: true, Sparse: true, Keys: bson.D{{Key: "email", Value: 1}}},
		{Collection: "userauth", Name: "userauth_email_unique", Unique: true, Sparse: true, Keys: bson.D{{Key: "email", Value: 1}}},
	} {
		if err := mongodb.EnsureIndex(ctx, client, db, spec); err != nil {
			t.Fatalf("unique index %s should build after migration: %v", spec.Name, err)
		}
	}

	// 幂等：再跑一次结果一致
	if err := Run(ctx, client, db, true, &bytes.Buffer{}); err != nil {
		t.Fatalf("second apply should be idempotent: %v", err)
	}
	if e, _ := getEmail(t, ui, oidNew); e != "12345678@qq.com" {
		t.Fatalf("keeper email changed after second run: %q", e)
	}
}

// userauth 自身存在重复（孤儿/不一致），且 userinfo 侧并无重复邮箱时，也必须收敛，
// 否则 userauth 唯一索引建不起来（H1）。
func TestRun_ResolvesUserauthOnlyDuplicate(t *testing.T) {
	client, db := testClient(t)
	ctx := context.Background()
	ui := client.Database(db).Collection("userinfo")
	ua := client.Database(db).Collection("userauth")

	keepInfo := primitive.NewObjectIDFromTimestamp(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
	orphanInfo := primitive.NewObjectIDFromTimestamp(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
	// userinfo 侧：一个持有邮箱，另一个无邮箱（模拟已摘/不一致）
	mustInsert(t, ui,
		bson.M{"_id": keepInfo, "email": "dup@qq.com"},
		bson.M{"_id": orphanInfo, "student_id": "S9"},
	)
	// userauth 侧：两条同邮箱，其中一条指向无邮箱的 userinfo
	mustInsert(t, ua,
		bson.M{"_id": primitive.NewObjectID(), "userInfoID": keepInfo, "email": "dup@qq.com"},
		bson.M{"_id": primitive.NewObjectID(), "userInfoID": orphanInfo, "email": "Dup@QQ.com"},
	)

	if err := Run(ctx, client, db, true, &bytes.Buffer{}); err != nil {
		t.Fatalf("apply: %v", err)
	}

	for _, spec := range []mongodb.IndexSpec{
		{Collection: "userinfo", Name: "userinfo_email_unique", Unique: true, Sparse: true, Keys: bson.D{{Key: "email", Value: 1}}},
		{Collection: "userauth", Name: "userauth_email_unique", Unique: true, Sparse: true, Keys: bson.D{{Key: "email", Value: 1}}},
	} {
		if err := mongodb.EnsureIndex(ctx, client, db, spec); err != nil {
			t.Fatalf("unique index %s should build: %v", spec.Name, err)
		}
	}
}

// 治理后若仍残留规范化，verifyNoDuplicates 必须报错（阻止盲目重启导致建索引 panic）。
func TestVerifyNoDuplicates_ReportsResidual(t *testing.T) {
	client, db := testClient(t)
	ui := client.Database(db).Collection("userinfo")
	mustInsert(t, ui,
		bson.M{"_id": primitive.NewObjectID(), "email": "x@qq.com"},
		bson.M{"_id": primitive.NewObjectID(), "email": "x@qq.com"},
	)
	if err := verifyNoDuplicates(context.Background(), ui, client.Database(db).Collection("userauth"), &bytes.Buffer{}); err == nil {
		t.Fatal("verifyNoDuplicates must fail when duplicates remain")
	}
}

// 纯空白邮箱会被规范化为空串，空串在 sparse 唯一索引下仍参与唯一性 -> 必须摘除。
func TestRun_UnsetsWhitespaceOnlyEmail(t *testing.T) {
	client, db := testClient(t)
	ui := client.Database(db).Collection("userinfo")
	id := primitive.NewObjectID()
	mustInsert(t, ui, bson.M{"_id": id, "email": "   "})

	if err := Run(context.Background(), client, db, true, &bytes.Buffer{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, ok := getEmail(t, ui, id); ok {
		t.Fatal("whitespace-only email must be unset")
	}
}

// 多条显式空串不参与 scanEmailGroups，但会撞 sparse 唯一索引 -> 自检必须拒绝。
func TestVerifyNoDuplicates_RejectsMultipleEmpty(t *testing.T) {
	client, db := testClient(t)
	ui := client.Database(db).Collection("userinfo")
	mustInsert(t, ui,
		bson.M{"_id": primitive.NewObjectID(), "email": ""},
		bson.M{"_id": primitive.NewObjectID(), "email": ""},
	)
	if err := verifyNoDuplicates(context.Background(), ui, client.Database(db).Collection("userauth"), &bytes.Buffer{}); err == nil {
		t.Fatal("verifyNoDuplicates must reject multiple empty emails")
	}
}
