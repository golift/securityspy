package securityspy

// Settings types for SecuritySpy v6 ++settings-* endpoints.
// Nested XML blocks (accounts, groups, storage paths) are omitted; use form values for those.

// GeneralSettings holds SecuritySpy v6 settings from settings-general-v6.20.xml.
type GeneralSettings struct {
	AllowSleep         YesNoBool `xml:"allowSleep"`
	AudioDeviceIndex   int       `xml:"audioDeviceIndex"`
	AudioDeviceName    string    `xml:"audioDeviceName"`
	AudioDeviceVol     int       `xml:"audioDeviceVol"`
	AudioDeviceVolText int       `xml:"audioDeviceVolText"`
	AutoReopen         YesNoBool `xml:"autoReopen"`
	DateFormat         int       `xml:"dateFormat"`
	DismissAlerts      YesNoBool `xml:"dismissAlerts"`
	ErrWindow          YesNoBool `xml:"errWindow"`
	FullVol            YesNoBool `xml:"fullVol"`
	HissReduction      YesNoBool `xml:"hissReduction"`
	ImageShare         YesNoBool `xml:"imageShare"`
	MingSave           YesNoBool `xml:"mingSave"`
	MuteIncoming       YesNoBool `xml:"muteIncoming"`
	SendDiagnostics    YesNoBool `xml:"sendDiagnostics"`
	SharePermission    int       `xml:"sharePermission"`
	SuspendDecoding    YesNoBool `xml:"suspendDecoding"`
	SysName            string    `xml:"sysName"`
	ThumbCrop          int       `xml:"thumbCrop"`
	TimeFormat         int       `xml:"timeFormat"`
	UpdateNotify       int       `xml:"updateNotify"`
}

// DisplaySettings holds SecuritySpy v6 settings from settings-display-v6.20.xml.
type DisplaySettings struct {
	AutoClose             int       `xml:"autoClose"`
	AutoCloseMins         int       `xml:"autoCloseMins"`
	CropMode              int       `xml:"cropMode"`
	DefaultWindow         int       `xml:"defaultWindow"`
	DisplayQuality        int       `xml:"displayQuality"`
	DivThickness          int       `xml:"divThickness"`
	ExcludedScreens       string    `xml:"excludedScreens"`
	FloatVideo            YesNoBool `xml:"floatVideo"`
	InfoAudio             YesNoBool `xml:"infoAudio"`
	InfoBar               int       `xml:"infoBar"`
	InfoFps1              YesNoBool `xml:"infoFps1"`
	InfoFps2              YesNoBool `xml:"infoFps2"`
	InfoName              YesNoBool `xml:"infoName"`
	InfoStatus            YesNoBool `xml:"infoStatus"`
	KioskMode             YesNoBool `xml:"kioskMode"`
	LowRate               YesNoBool `xml:"lowRate"`
	MotionBox             YesNoBool `xml:"motionBox"`
	RememberCamControlPos YesNoBool `xml:"rememberCamControlPos"`
	ReplaySeconds         int       `xml:"replaySeconds"`
}

// StorageSettings holds SecuritySpy v6 settings from settings-storage-v6.20.xml.
type StorageSettings struct {
	ArchiveMode      int       `xml:"archiveMode"`
	ArchiveStorage   string    `xml:"archiveStorage"`
	DiskWaitTime     int       `xml:"diskWaitTime"`
	GlobalStorage    string    `xml:"globalStorage"`
	RemoveAgeNonSys  int       `xml:"removeAgeNonSys"`
	RemoveAgeSys     int       `xml:"removeAgeSys"`
	RemoveAutoNonSys int       `xml:"removeAutoNonSys"`
	RemoveAutoSys    int       `xml:"removeAutoSys"`
	RemoveByAge      YesNoBool `xml:"removeByAge"`
	RemoveBySpace    YesNoBool `xml:"removeBySpace"`
	RemoveGbNonSys   int       `xml:"removeGbNonSys"`
	RemoveGbSys      int       `xml:"removeGbSys"`
	Tag1             YesNoBool `xml:"tag-1"`
	Tag2             YesNoBool `xml:"tag-2"`
	Tag3             YesNoBool `xml:"tag-3"`
	Tag4             YesNoBool `xml:"tag-4"`
	Tag5             YesNoBool `xml:"tag-5"`
	Tag6             YesNoBool `xml:"tag-6"`
	Tag7             YesNoBool `xml:"tag-7"`
	UsageWarningGb   int       `xml:"usageWarningGb"`
}

// CompressionSettings holds SecuritySpy v6 settings from settings-compression-v6.20.xml.
type CompressionSettings struct {
	AudioCodec       int    `xml:"audioCodec"`
	AudioCodecName   string `xml:"audioCodecName"`
	AudioQuality     int    `xml:"audioQuality"`
	AudioQualityText int    `xml:"audioQualityText"`
	JpegQuality      int    `xml:"jpegQuality"`
	JpegQualityText  int    `xml:"jpegQualityText"`
	VideoCodec       int    `xml:"videoCodec"`
	VideoCodecName   string `xml:"videoCodecName"`
	VideoQuality     int    `xml:"videoQuality"`
	VideoQualityText int    `xml:"videoQualityText"`
}

// EmailSettings holds SecuritySpy v6 settings from settings-email-v6.20.xml.
type EmailSettings struct {
	Address       string    `xml:"address"`
	DowntimeEmail YesNoBool `xml:"downtimeEmail"`
	Encryption    int       `xml:"encryption"`
	ErrEmailLevel int       `xml:"errEmailLevel"`
	Fps           int       `xml:"fps"`
	FromEmail     string    `xml:"fromEmail"`
	FromName      string    `xml:"fromName"`
	ImageCount    int       `xml:"imageCount"`
	MaxRes        int       `xml:"maxRes"`
	MediaType     int       `xml:"mediaType"`
	SendingMethod int       `xml:"sendingMethod"`
	StatsEmail    YesNoBool `xml:"statsEmail"`
	Subject       string    `xml:"subject"`
	SysEmail      string    `xml:"sysEmail"`
	Username      string    `xml:"username"`
}

// WebSettings holds SecuritySpy v6 settings from settings-web-v6.20.xml.
type WebSettings struct {
	AutoNatHTTP      YesNoBool `xml:"autoNatHttp"`
	AutoNatHTTPS     YesNoBool `xml:"autoNatHttps"`
	Bonjour          YesNoBool `xml:"bonjour"`
	CorsDomains      string    `xml:"corsDomains"`
	DdnsName         string    `xml:"ddnsName"`
	DdnsStatus       int       `xml:"ddnsStatus"`
	GeoblockList     string    `xml:"geoblockList"`
	GeoblockType     int       `xml:"geoblockType"`
	HlsMaxFps        int       `xml:"hlsMaxFps"`
	HlsMaxRes        int       `xml:"hlsMaxRes"`
	HTTP             YesNoBool `xml:"http"`
	HTTPS            YesNoBool `xml:"https"`
	Iframe           YesNoBool `xml:"iframe"`
	Legacy           YesNoBool `xml:"legacy"`
	ListenIps        string    `xml:"listenIps"`
	Log              YesNoBool `xml:"log"`
	NoHeif           YesNoBool `xml:"noHeif"`
	PortHTTP         int       `xml:"portHttp"`
	PortHTTPWan      int       `xml:"portHttpWan"`
	PortHTTPS        int       `xml:"portHttps"`
	PortHTTPSWan     int       `xml:"portHttpsWan"`
	PublicResources  string    `xml:"publicResources"`
	ScreenControl    YesNoBool `xml:"screenControl"`
	SessionLen       int       `xml:"sessionLen"`
	UserHeaders      string    `xml:"userHeaders"`
	UserWanDetails   YesNoBool `xml:"userWanDetails"`
	VariableFps      YesNoBool `xml:"variableFps"`
	VideoPassthrough YesNoBool `xml:"videoPassthrough"`
	WanAddress       string    `xml:"wanAddress"`
}

// CameraSettings holds SecuritySpy v6 settings from settings-cameras-v6.20.xml.
type CameraSettings struct {
	AComeFront             YesNoBool `xml:"aComeFront"`
	ADelay                 int       `xml:"aDelay"`
	AEmail                 string    `xml:"aEmail"`
	ANotification          YesNoBool `xml:"aNotification"`
	ARedBox                YesNoBool `xml:"aRedBox"`
	ARedBoxDuration        int       `xml:"aRedBoxDuration"`
	AReset                 int       `xml:"aReset"`
	AResetType             int       `xml:"aResetType"`
	AShellCommand          string    `xml:"aShellCommand"`
	ASoundDurationCam      int       `xml:"aSoundDurationCam"`
	ASoundDurationMac      int       `xml:"aSoundDurationMac"`
	ASoundVolCam           int       `xml:"aSoundVolCam"`
	ASoundVolMac           int       `xml:"aSoundVolMac"`
	ATriggerAudio          YesNoBool `xml:"aTriggerAudio"`
	ATriggerCamMd          YesNoBool `xml:"aTriggerCamMd"`
	ATriggerCamP1          YesNoBool `xml:"aTriggerCamP1"`
	ATriggerCamP2          YesNoBool `xml:"aTriggerCamP2"`
	ATriggerCamPir         YesNoBool `xml:"aTriggerCamPir"`
	ATriggerHome           YesNoBool `xml:"aTriggerHome"`
	ATriggerMotion         YesNoBool `xml:"aTriggerMotion"`
	ATriggerMotionA        YesNoBool `xml:"aTriggerMotionA"`
	ATriggerMotionH        YesNoBool `xml:"aTriggerMotionH"`
	ATriggerMotionV        YesNoBool `xml:"aTriggerMotionV"`
	AVolTextCam            int       `xml:"aVolTextCam"`
	AVolTextMac            int       `xml:"aVolTextMac"`
	AWakeScreen            YesNoBool `xml:"aWakeScreen"`
	Address                string    `xml:"address"`
	AnimalBird             YesNoBool `xml:"animalBird"`
	AnimalFish             YesNoBool `xml:"animalFish"`
	AnimalQuadruped        YesNoBool `xml:"animalQuadruped"`
	AnimalSensitivity      int       `xml:"animalSensitivity"`
	AnimalSensitivityText  int       `xml:"animalSensitivityText"`
	AudioDeviceVol         int       `xml:"audioDeviceVol"`
	AudioSensitivity       int       `xml:"audioSensitivity"`
	AudioSensitivityText   int       `xml:"audioSensitivityText"`
	Brightness             int       `xml:"brightness"`
	CameraNum              int       `xml:"cameraNum"`
	CCFreq                 int       `xml:"ccFreq"`
	CCImage                YesNoBool `xml:"ccImage"`
	CCImageInterval        int       `xml:"ccImageInterval"`
	CCMovie                YesNoBool `xml:"ccMovie"`
	CCMovieFps             int       `xml:"ccMovieFps"`
	CCMoviePlaybackFps     int       `xml:"ccMoviePlaybackFps"`
	CCRemoveAge            int       `xml:"ccRemoveAge"`
	Cmd0                   string    `xml:"cmd0"`
	Cmd1                   string    `xml:"cmd1"`
	Cmd2                   string    `xml:"cmd2"`
	Cmd3                   string    `xml:"cmd3"`
	CmdName0               string    `xml:"cmdName0"`
	CmdName1               string    `xml:"cmdName1"`
	CmdName2               string    `xml:"cmdName2"`
	CmdName3               string    `xml:"cmdName3"`
	ConfigVital            int       `xml:"configVital"`
	ConfigureHome          int       `xml:"configureHome"`
	Contrast               int       `xml:"contrast"`
	DeviceType             int       `xml:"deviceType"`
	Enabled                YesNoBool `xml:"enabled"`
	Fps                    int       `xml:"fps"`
	FrameEndMethod         int       `xml:"frameEndMethod"`
	Height                 int       `xml:"height"`
	HomeShortcut0          int       `xml:"homeShortcut0"`
	HomeShortcut1          int       `xml:"homeShortcut1"`
	HomeShortcut2          int       `xml:"homeShortcut2"`
	HomeShortcut3          int       `xml:"homeShortcut3"`
	HomeShortcut4          int       `xml:"homeShortcut4"`
	HomeShortcut5          int       `xml:"homeShortcut5"`
	HomeShortcut6          int       `xml:"homeShortcut6"`
	HomeShortcut7          int       `xml:"homeShortcut7"`
	HumanSensitivity       int       `xml:"humanSensitivity"`
	HumanSensitivityText   int       `xml:"humanSensitivityText"`
	IntPresets             YesNoBool `xml:"intPresets"`
	InvertPanTilt          YesNoBool `xml:"invertPanTilt"`
	LastFormat             int       `xml:"lastFormat"`
	LastInputNum           int       `xml:"lastInputNum"`
	MCDaily                int       `xml:"mcDaily"`
	MCImage                YesNoBool `xml:"mcImage"`
	MCImageInterval        int       `xml:"mcImageInterval"`
	MCImagePost            int       `xml:"mcImagePost"`
	MCMovie                YesNoBool `xml:"mcMovie"`
	MCMovieFps             int       `xml:"mcMovieFps"`
	MCMoviePost            int       `xml:"mcMoviePost"`
	MCMoviePre             int       `xml:"mcMoviePre"`
	MCRemoveAge            int       `xml:"mcRemoveAge"`
	MCTriggerAudio         YesNoBool `xml:"mcTriggerAudio"`
	MCTriggerCamMd         YesNoBool `xml:"mcTriggerCamMd"`
	MCTriggerCamP1         YesNoBool `xml:"mcTriggerCamP1"`
	MCTriggerCamP2         YesNoBool `xml:"mcTriggerCamP2"`
	MCTriggerCamPir        YesNoBool `xml:"mcTriggerCamPir"`
	MCTriggerHome          YesNoBool `xml:"mcTriggerHome"`
	MCTriggerMotion        YesNoBool `xml:"mcTriggerMotion"`
	MCTriggerMotionA       YesNoBool `xml:"mcTriggerMotionA"`
	MCTriggerMotionH       YesNoBool `xml:"mcTriggerMotionH"`
	MCTriggerMotionV       YesNoBool `xml:"mcTriggerMotionV"`
	MdType                 int       `xml:"mdType"`
	MotionMask             string    `xml:"motionMask"`
	MotionSensitivity      int       `xml:"motionSensitivity"`
	MotionSensitivityText  int       `xml:"motionSensitivityText"`
	Name                   string    `xml:"name"`
	NoAudioSend            YesNoBool `xml:"noAudioSend"`
	NoPtz                  YesNoBool `xml:"noPtz"`
	OmitOptions            YesNoBool `xml:"omitOptions"`
	OverlayPos             int       `xml:"overlayPos"`
	OverlaySize            int       `xml:"overlaySize"`
	OverlayText            string    `xml:"overlayText"`
	Pasp                   YesNoBool `xml:"pasp"`
	Password               string    `xml:"password"`
	PermissiveSSL          YesNoBool `xml:"permissiveSsl"`
	PortHTTP               int       `xml:"portHttp"`
	PortRTSP               int       `xml:"portRtsp"`
	PrivacyMask            string    `xml:"privacyMask"`
	PtzMdWait              int       `xml:"ptzMdWait"`
	Quality                int       `xml:"quality"`
	RecompressAudio        YesNoBool `xml:"recompressAudio"`
	RecompressVideo        YesNoBool `xml:"recompressVideo"`
	Request                string    `xml:"request"`
	SetSockBuffer          YesNoBool `xml:"setSockBuffer"`
	SetTime                YesNoBool `xml:"setTime"`
	SslHTTP                YesNoBool `xml:"sslHttp"`
	SslRTSP                YesNoBool `xml:"sslRtsp"`
	SuppressErr            YesNoBool `xml:"suppressErr"`
	SwTimestamps           YesNoBool `xml:"swTimestamps"`
	Transformation         int       `xml:"transformation"`
	Username               string    `xml:"username"`
	VehicleSensitivity     int       `xml:"vehicleSensitivity"`
	VehicleSensitivityText int       `xml:"vehicleSensitivityText"`
	ViewOnly               YesNoBool `xml:"viewOnly"`
	WebcamFreq             int       `xml:"webcamFreq"`
	WebcamName             string    `xml:"webcamName"`
	Width                  int       `xml:"width"`
}
