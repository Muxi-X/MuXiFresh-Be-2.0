// Code generated by goctl. DO NOT EDIT.
package types

type ClickReq struct {
	Authorization string `header:"Authorization"`
}

type ClickResp struct {
	UserType   string `json:"user_type"`
	FormStatus string `json:"form_status"`
}

type CreateReq struct {
	Authorization string `header:"Authorization"`
	FormId        string `json:"form_id,optional"`
	Avatar        string `json:"avatar"`
	Major         string `json:"major"`
	Grade         string `json:"grade"`
	Gender        string `json:"gender"`
	Phone         string `json:"phone"`
	Group         string `json:"group,options=[Product,Design,Frontend,Backend,Android]"`
	Reason        string `json:"reason"`
	Knowledge     string `json:"knowledge"`
	SelfIntro     string `json:"self_intro"`
	ExtraQuestion string `json:"extra_question"`
}

type CreateResp struct {
	Flag bool `json:"flag"`
}

type CheckReq struct {
	Authorization string `header:"Authorization"`
	EntryFormID   string `form:"entry_form_id"`
}

type CheckResp struct {
	FormId        string `json:"form_id"`
	Avatar        string `json:"avatar"`
	Major         string `json:"major"`
	Grade         string `json:"grade"`
	Gender        string `json:"gender"`
	Phone         string `json:"phone"`
	Group         string `json:"group"`
	Reason        string `json:"reason"`
	Knowledge     string `json:"knowledge"`
	SelfIntro     string `json:"self_intro"`
	ExtraQuestion string `json:"extra_question"`
}

type GetApplicantNumberReq struct {
	Authorization string `header:"Authorization"`
	Group         string `json:"group,options=[Product,Design,Frontend,Backend,Android]"`
	Year          int64  `json:"year"`
	Season        string `json:"season,options=[autumn,spring]"`
}

type GetApplicantNumberResp struct {
	Number int64 `json:"number"`
}
