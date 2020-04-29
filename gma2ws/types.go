package gma2ws

type RequestType string
type ItemType int
type ExecButtonViewMode int

const (
	RequestTypeGetData            RequestType = "getdata"
	RequestTypeLogin              RequestType = "login"
	RequestTypePlaybacks          RequestType = "playbacks"
	RequestTypeClose              RequestType = "close"
	RequestTypePlaybacksUserInput RequestType = "playbacks_userInput"
	RequestTypeKeyname            RequestType = "keyname"
	RequestTypeCommand            RequestType = "command"

	ExecButtonViewModeFader ExecButtonViewMode = 1
)

type ServerStatus struct {
	Status  string `json:"status"`
	AppType string `json:"appType"`
}

type ClientHanshake struct {
	Session int `json:"session"`
}

type ServerLoginParams struct {
	Realtime   bool `json:"realtime"`
	Session    int  `json:"session"`
	ForceLogin bool `json:"forceLogin"`
	WorldIndex int  `json:"worldIndex"`
}

type ServerLoginResponse struct {
	Realtime    bool   `json:"realtime"`
	RequestType string `json:"requestType"`
	Result      bool   `json:"result"`
	WorldIndex  int    `json:"worldIndex"`
}

type ClientRequestGetData struct {
	ClientRequestGeneric
	Data string `json:"data"`
}

type ClientRequestLogin struct {
	ClientRequestGeneric
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientRequestGeneric struct {
	RequestType RequestType `json:"requestType"`
	Session     int         `json:"session"`
	MaxRequests int         `json:"maxRequests"`
}

type ClientRequestPlaybacks struct {
	ClientRequestGeneric
	PageIndex          int                `json:"pageIndex"`  // Page selected
	StartIndex         []int              `json:"startIndex"` // For the different ranges: executors to select (start index)
	ItemsCount         []int              `json:"itemsCount"` // For the different ranges: executors to select (count)
	ItemsType          []int              `json:"itemsType"`  // ???
	View               int                `json:"view"`       // ???
	ExecButtonViewMode ExecButtonViewMode `json:"execButtonViewMode"`
	ButtonsViewMode    int                `json:"buttonsViewMode"` // ???
}

type ServerResponseGeneric struct {
	ResponseType RequestType `json:"responseType"`
	Realtime     bool        `json:"realtime"`
	WorldIndex   int         `json:"worldIndex"`
}

type ServerResponseGetData struct {
	ServerResponseGeneric
	Data []map[string]string `json:"data"`
}

type ServerResponsePlayback struct {
	ServerResponseGeneric
	ResponseSubType int               `json:"responseSubType"` // ???
	IPage           int               `json:"iPage"`           // ???
	ItemGroups      []ServerPlaybacks `json:"itemGroups"`      // 1 ServerPlaybacks per requested ranges
}

type ServerPlaybacks struct {
	ItemsType int                `json:"itemsType"` // ???
	CntPages  int                `json:"cntPages"`  // ???
	Items     [][]ServerPlayback `json:"items"`     // Items per fader groups (fader groups are 5 fader wide)
}

type ServerPlayback struct {
	Index                 ServerPlaybackTextItem         `json:"i"`
	ObjectType            ServerPlaybackTextItem         `json:"oType"`
	ObjectIndex           ServerPlaybackTextItem         `json:"oI"`
	TextTop               ServerPlaybackTextItem         `json:"tt"`
	HeaderBackgroundColor string                         `json:"bC"`
	HeaderBorderColor     string                         `json:"bdC"`
	Cues                  ServerPlaybackCues             `json:"cues"`
	CombinedItems         int                            `json:"combinedItems"`
	IsExec                int                            `json:"iExec"`
	IsRun                 int                            `json:"isRun"`
	ExecutorBlocks        []ServerPlaybackExecutorBlocks `json:"executorBlocks"` // Why an array here ???
}

type ServerPlaybackTextItem struct {
	Text     string                  `json:"t"`
	Color    string                  `json:"c"`
	Progress *ServerPlaybackProgress `json:"pgs"`
}

type ServerPlaybackProgress struct {
	Value           float64 `json:"v"`
	BackgroundColor string  `json:"bC"`
}

type ServerPlaybackCues struct {
	BackgroundColor string                   `json:"bC"`
	Items           []ServerPlaybackTextItem `json:"items"`
}

type ServerPlaybackExecutorBlocks struct {
	Button1 *ServerPlaybackExecutorBlock `json:"button1"`
	Button2 *ServerPlaybackExecutorBlock `json:"button2"`
	Button3 *ServerPlaybackExecutorBlock `json:"button3"`
	Fader   *ServerPlaybackExecutorBlock `json:"fader"`
}

type ServerPlaybackExecutorBlock struct {
	ID          int         `json:"id"`
	Text        string      `json:"t"`
	S           bool        `json:"s"` // ???
	Color       string      `json:"color"`
	BorderColor string      `json:"bdC"`
	LeftLed     interface{} `json:"leftLED"`  // Seems to never be filled
	RightLed    interface{} `json:"rightLED"` // Same
	TypeText    string      `json:"tt"`       // Only on faders
	Value       float64     `json:"v"`        // Only on faders
	ValueText   string      `json:"vT"`       // Only on faders
	Min         float64     `json:"min"`      // Only on faders
	Max         float64     `json:"max"`      // Only on faders
}

type ClientRequestPlaybacksUserInput struct {
	ClientRequestGeneric
	Executor int     `json:"execIndex"`
	Page     int     `json:"pageIndex"`
	Value    float64 `json:"faderValue"`
	Type     int     `json:"type"`
	ButtonID int     `json:"buttonId"`
	Pressed  bool    `json:"pressed"`
	Released bool    `json:"released"`
}

type ClientRequestKeyName struct {
	ClientRequestGeneric
	KeyName         KeyName `json:"keyname"`
	Value           int     `json:"value"`
	CommandLineText string  `json:"cmdlinetext"`
	AutoSubmit      bool    `json:"autoSubmit"`
}

type ClientRequestCommand struct {
	ClientRequestGeneric
	Command string `json:"command"`
}
