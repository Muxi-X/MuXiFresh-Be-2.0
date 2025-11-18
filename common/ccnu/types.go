package ccnusvc

import (
	"context"
	"github.com/MuxiKeStack/be-ccnu/domain"
	"time"
)

type CCNUService interface {
	Login(ctx context.Context, studentId string, password string) (bool, error)
	GetSelfCourseList(ctx context.Context, studentId, password, year, term string) ([]domain.Course, error)
	// GetSelfGradeList 这个是只能获取总分，没有聚合平时成绩等细节，现在主要用于准确获取个人历史课程
	GetSelfGradeList(ctx context.Context, studentId, password, year, term string) ([]domain.Grade, error)
	// GetAllDetailOfGrade 获取所有成绩的所有细节
	GetDetailOfGradeList(ctx context.Context, studentId string, password string, year string, term string) ([]domain.Grade, error)
}

type ccnuService struct {
	timeout time.Duration
}

func NewCCNUService() CCNUService {
	return &ccnuService{
		timeout: time.Second * 5,
	}
}
