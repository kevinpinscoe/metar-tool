# metar-tool

Is a command line tool written in Go used for dumping the latest METAR observations either as a text product or as JSON. 

You can also obtain the latest forecast from a US NWS Weather Forecast Office (WFO). 

You can take the output and pipe or redirect it to a file for archiving.

This is the initial draft. The next version will allow you to decode a METAR
string to a full text description. This will allow you to run the tool against
an archived or saved JSON file and decode it. The decoding will work on live
data as well. You will even be able to `metar-tool --obs ktys --json | metar-tool --decode`.

## Usage examples

```
metar-tool --obs ktys
METAR KTYS 142253Z 21012KT 10SM BKN026 OVC034 07/04 A2968 RMK AO2 RAE04 SLP049 P0001 T00670039

metar-tool --obs ktys --json
[{"icaoId":"KTYS","receiptTime":"2026-01-14T22:56:53.763Z","obsTime":1768431180,"reportTime":"2026-01-14T23:00:00.000Z","temp":6.7,"dewp":3.9,"wdir":210,"wspd":12,"visib":"10+","altim":1005.2,"slp":1004.9,"qcField":4,"precip":0.01,"metarType":"METAR","rawOb":"METAR KTYS 142253Z 21012KT 10SM BKN026 OVC034 07/04 A2968 RMK AO2 RAE04 SLP049 P0001 T00670039","lat":35.818,"lon":-83.9857,"elev":300,"name":"Knoxville/Tyson Arpt, TN, US","cover":"OVC","clouds":[{"cover":"BKN","base":2600},{"cover":"OVC","base":3400}],"fltCat":"MVFR"}]

metar-tool --obs ktys -json -pretty
[
  {
    "altim": 1014.3,
    "clouds": [
      {
        "base": 2300,
        "cover": "FEW"
      },
      {
        "base": 4000,
        "cover": "FEW"
      }
    ],
    "cover": "FEW",
    "dewp": -13.9,
    "elev": 300,
    "fltCat": "VFR",
    "icaoId": "KTYS",
    "lat": 35.818,
    "lon": -83.9857,
    "metarType": "METAR",
    "name": "Knoxville/Tyson Arpt, TN, US",
    "obsTime": 1768481580,
    "qcField": 4,
    "rawOb": "METAR KTYS 151253Z 31007KT 10SM FEW023 FEW040 M06/M14 A2995 RMK AO2 SLP148 T10561139",
    "receiptTime": "2026-01-15T12:56:56.412Z",
    "reportTime": "2026-01-15T13:00:00.000Z",
    "slp": 1014.8,
    "temp": -5.6,
    "visib": "10+",
    "wdir": 310,
    "wspd": 7
  }
]

metar-tool --forecast nws kmrx
MRX (Area Forecast Discussion) - issued 
------------------------------------------------------------------------

000
FXUS64 KMRX 141802
AFDMRX

Area Forecast Discussion
National Weather Service Morristown TN
102 PM EST Wed Jan 14 2026

...New DISCUSSION, AVIATION...

.KEY MESSAGES...
Updated at 100 PM EST Wed Jan 14 2026

- Significant accumulating snow is expected across the higher 
  elevations with lighter accumulations possible across portions 
  of the Plateau and Valley from late this afternoon and evening 
  through early Thursday morning. 
 ...

# To archive an observation in local time
 metar-tool --obs ktys --output "ktys-$(date +%Y%m%d-%H%M).txt"

# To archive an observation in UTC
 metar-tool --obs ktys --output "ktys-$(date -u +%Y%m%d-%H%MZ).txt"

 # Add --verbose to have file name announced
 metar-tool --obs KTYS --output "ktys-$(date -u +%Y%m%d-%H%MZ).txt" --verbose
 Writing output to ktys-20260115-1337Z.txt
```

## More about METAR

METAR stands for METeorological Aerodrome Report. METAR is a format for weather reporting that is predominately used for pilots and meteorologists. These reports are issued at each reporting location every hour and are considered valid weather information for 1 hour.

You can read more at https://www.weather.gov/asos/METAR.html 

METAR abbreviations can be found at https://www.weather.gov/media/wrh/mesowest/metar_decode_key.pdf


## List of WFO's 

National Weather Service Weather Forecast Office (WFO), one of 122 local offices in the 
U.S. responsible for issuing timely, localized weather forecasts, watches, and warnings (like for severe storms or floods) for specific regions (County Warning Areas) to protect life and property, staffed by meteorologists 24/7

A list of WFO's with their designated codes can be found at https://www.weather.gov/nwr/wfo_nwr 


## Install

`make install` by default will install into $HOME/.local.

To override this use something like `sudo make install PREFIX=/usr/local`

## Release

A Github Actions is planned to create and tag releases for Linux, Mac, Raspberry Pi (ARM) and Windows 