package consts

type CopySetting string

const (
	CopySettingNone     CopySetting = ""         // 无限制
	CopySettingAppend   CopySetting = "append"   // 增加内容尾巴
	CopySettingDisabled CopySetting = "disabled" // 禁止复制内容
)

type WatermarkSetting string

const (
	WatermarkDisabled WatermarkSetting = ""        // 未开启水印
	WatermarkHidden   WatermarkSetting = "hidden"  // 隐形水印
	WatermarkVisible  WatermarkSetting = "visible" // 显性水印
)

type HomePageSetting string

const (
	HomePageSettingDoc    HomePageSetting = "doc"    // 文档页面
	HomePageSettingCustom HomePageSetting = "custom" // 自定义首页
)
