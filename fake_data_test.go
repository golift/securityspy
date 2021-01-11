package securityspy

// Test data for Test methods.
// This is all copied directly from ++systemInfo, ++scripts and ++sounds

const (
	testServerInfo = `<server>
	<name>SecuritySpy</name>
	<version>4.2.10</version>
	<uuid>C02L1333A8J2FkXIZC2O</uuid>
	<eventstreamcount>4206</eventstreamcount>
	<ddns-name></ddns-name>
	<wan-address>server.host.name</wan-address>
	<server-name></server-name>
	<bonjour-name></bonjour-name>
	<ip1>192.168.1.1</ip1>
	<ip2>192.168.2.1</ip2>
	<http-enabled>yes</http-enabled>
	<http-port>8000</http-port>
	<http-port-wan>80</http-port-wan>
	<https-enabled>no</https-enabled>
	<https-port>8001</https-port>
	<https-port-wan>8001</https-port-wan>
	<current-local-time>2019-02-10T15:53:23-08:00</current-local-time>
	<seconds-from-gmt>-28800</seconds-from-gmt>
	<date-format>MM/DD/YYYY</date-format>
	<time-format>24</time-format>
</server>`

	testCameraOne = `<camera>
	<number>1</number>
	<connected>yes</connected>
	<width>2304</width>
	<height>1296</height>
	<mode>active</mode>
	<mode-c>armed</mode-c>
	<mode-m>armed</mode-m>
	<mode-a>armed</mode-a>
	<hasaudio>yes</hasaudio>
	<ptzcapabilities>0</ptzcapabilities>
	<timesincelastframe>0</timesincelastframe>
	<timesincelastmotion>7985</timesincelastmotion>
	<devicename>ONVIF</devicename>
	<devicetype>Network</devicetype>
	<address>192.168.1.12</address>
	<port></port>
	<port-rtsp></port-rtsp>
	<request></request>
	<name>Porch</name>
	<overlay>no</overlay>
	<overlaytext>+d</overlaytext>
	<transformation>0</transformation>
	<audio_network>yes</audio_network>
	<audio_devicename></audio_devicename>
	<md_enabled>yes</md_enabled>
	<md_sensitivity>51</md_sensitivity>
	<md_triggertime_x2>2</md_triggertime_x2>
	<md_capture>yes</md_capture>
	<md_capturefps>20</md_capturefps>
	<md_precapture>3</md_precapture>
	<md_postcapture>10</md_postcapture>
	<md_captureimages>no</md_captureimages>
	<md_uploadimages>no</md_uploadimages>
	<md_recordaudio>yes</md_recordaudio>
	<md_audiotrigger>no</md_audiotrigger>
	<md_audiothreshold>50</md_audiothreshold>
	<action_scriptname>SS_SendiMessages.scpt</action_scriptname>
	<action_soundname></action_soundname>
	<action_resettime>60</action_resettime>
	<tl_capture>no</tl_capture>
	<tl_recordaudio>yes</tl_recordaudio>
	<current-fps>20.202</current-fps>
	<schedule-id-cc>1</schedule-id-cc>
	<schedule-id-mc>2</schedule-id-mc>
	<schedule-id-a>3</schedule-id-a>
	<schedule-override-cc>0</schedule-override-cc>
	<schedule-override-mc>1</schedule-override-mc>
	<schedule-override-a>2</schedule-override-a>
	<preset-name-1></preset-name-1>
	<preset-name-2></preset-name-2>
	<preset-name-3></preset-name-3>
	<preset-name-4></preset-name-4>
	<preset-name-5></preset-name-5>
	<preset-name-6></preset-name-6>
	<preset-name-7></preset-name-7>
	<preset-name-8></preset-name-8>
	<permissions>63167</permissions>
</camera>`

	testCameraTwo = `<camera>
	<number>2</number>
	<connected>yes</connected>
	<width>2592</width>
	<height>1520</height>
	<mode>passive</mode>
	<mode-c>armed</mode-c>
	<mode-m>armed</mode-m>
	<mode-a>disarmed</mode-a>
	<hasaudio>no</hasaudio>
	<ptzcapabilities>31</ptzcapabilities>
	<timesincelastframe>0</timesincelastframe>
	<timesincelastmotion>4</timesincelastmotion>
	<devicename>ONVIF</devicename>
	<devicetype>Network</devicetype>
	<address>192.168.1.11</address>
	<port></port>
	<port-rtsp></port-rtsp>
	<request></request>
	<name>Road</name>
	<overlay>no</overlay>
	<overlaytext>+d</overlaytext>
	<transformation>0</transformation>
	<audio_network>no</audio_network>
	<audio_devicename></audio_devicename>
	<md_enabled>yes</md_enabled>
	<md_sensitivity>47</md_sensitivity>
	<md_triggertime_x2>1</md_triggertime_x2>
	<md_capture>yes</md_capture>
	<md_capturefps>20</md_capturefps>
	<md_precapture>3</md_precapture>
	<md_postcapture>5</md_postcapture>
	<md_captureimages>no</md_captureimages>
	<md_uploadimages>no</md_uploadimages>
	<md_recordaudio>yes</md_recordaudio>
	<md_audiotrigger>no</md_audiotrigger>
	<md_audiothreshold>50</md_audiothreshold>
	<action_scriptname>SS_SendiMessages.scpt</action_scriptname>
	<action_soundname></action_soundname>
	<action_resettime>59</action_resettime>
	<tl_capture>no</tl_capture>
	<tl_recordaudio>yes</tl_recordaudio>
	<current-fps>20.000</current-fps>
	<schedule-id-cc>1</schedule-id-cc>
	<schedule-id-mc>1</schedule-id-mc>
	<schedule-id-a>3</schedule-id-a>
	<schedule-override-cc>0</schedule-override-cc>
	<schedule-override-mc>0</schedule-override-mc>
	<schedule-override-a>0</schedule-override-a>
	<preset-name-1></preset-name-1>
	<preset-name-2></preset-name-2>
	<preset-name-3></preset-name-3>
	<preset-name-4></preset-name-4>
	<preset-name-5></preset-name-5>
	<preset-name-6></preset-name-6>
	<preset-name-7></preset-name-7>
	<preset-name-8></preset-name-8>
	<permissions>62975</permissions>
</camera>`

	testScheduleList = `<schedulelist>
  <schedule>
    <name>Unarmed 24/7</name>
    <id>0</id>
  </schedule>
  <schedule>
    <name>Armed 24/7</name>
    <id>1</id>
  </schedule>
  <schedule>
    <name>Armed Sunrise To Sunset</name>
    <id>2</id>
  </schedule>
  <schedule>
    <name>Armed Sunset To Sunrise</name>
    <id>3</id>
  </schedule>
  <schedule>
    <name>MyFirstSchedule</name>
    <id>4903</id>
  </schedule>
  <schedule>
    <name>AnotherSchedule</name>
    <id>1741</id>
  </schedule>
</schedulelist>`

	testSchedulePresetList = `<schedulepresetlist>
  <schedulepreset>
    <name>MyFirstPreset</name>
    <id>1930238093</id>
  </schedulepreset>
</schedulepresetlist>`

	testScheduleOverrideList = `<scheduleoverridelist>
  <scheduleoverride>
    <name>None</name>
    <id>0</id>
  </scheduleoverride>
  <scheduleoverride>
    <name>Unarmed Until Next Scheduled Event</name>
    <id>1</id>
  </scheduleoverride>
  <scheduleoverride>
    <name>Armed Until Next Scheduled Event</name>
    <id>2</id>
  </scheduleoverride>
  <scheduleoverride>
    <name>Unarmed For 1 Hour</name>
    <id>3</id>
  </scheduleoverride>
  <scheduleoverride>
    <name>Armed For 1 Hour</name>
    <id>4</id>
  </scheduleoverride>
</scheduleoverridelist>`

	testSystemInfo = `<?xml version="1.0" encoding="utf-8"?><system>` + testServerInfo +
		`<cameralist>` + testCameraOne + testCameraTwo + `</cameralist>` +
		testScheduleList + testScheduleOverrideList + testSchedulePresetList + `</system>`

	testSoundsList = `<?xml version="1.0" encoding="utf-8"?>
<sounds>
<name>Beeps.aif</name>
<name>Bell ring.aif</name>
<name>High pitch fast siren.aif</name>
<name>Human scream.aif</name>
<name>Medium pitch fast siren.aif</name>
<name>Rising siren.aif</name>
<name>Basso.aiff</name>
<name>Blow.aiff</name>
<name>Bottle.aiff</name>
<name>Frog.aiff</name>
<name>Funk.aiff</name>
<name>Glass.aiff</name>
<name>Hero.aiff</name>
<name>Morse.aiff</name>
<name>Ping.aiff</name>
<name>Pop.aiff</name>
<name>Purr.aiff</name>
<name>Sosumi.aiff</name>
<name>Submarine.aiff</name>
<name>Tink.aiff</name>
</sounds>`

	testScriptsList = `<?xml version="1.0" encoding="utf-8"?>
<scripts>
<name>Web-i Activate Relay 1.scpt</name>
<name>Web-i Activate Relay 2.scpt</name>
<name>Web-i Activate Relay 3.scpt</name>
<name>Web-i Activate Relay 4.scpt</name>
<name>Web-i Activate Relay 5.scpt</name>
<name>Web-i Activate Relay 6.scpt</name>
<name>Web-i Activate Relay 7.scpt</name>
<name>Web-i Activate Relay 8.scpt</name>
<name>WebRelay Activate Relay 1.scpt</name>
<name>WebRelay Activate Relay 2.scpt</name>
<name>WebRelay Activate Relay 3.scpt</name>
<name>WebRelay Activate Relay 4.scpt</name>
<name>WebRelay Activate Relay 5.scpt</name>
<name>WebRelay Activate Relay 6.scpt</name>
<name>WebRelay Activate Relay 7.scpt</name>
<name>WebRelay Activate Relay 8.scpt</name>
</scripts>`
)
