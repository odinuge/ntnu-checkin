# ntnu-checkin (-cli)

> Simple and dumb cli for ntnu checkin

*THIS IS NOT AFFILIATED WITH NTNU IN ANY WAY, AND COMES WITH NO SUPPORT*

(I do however think it is better that people check in via the CLI instead of _not_ doing it..)


```bash
$ go get github.com/odinuge/ntnu-checkin
$ export FEIDE_USERNAME=user
$ export FEIDE_PASSWORD=pass

$ ntnu-checkin search gamle elektro
ROOM-ID    NAME
3066       Sanntidssalen Gamle elektro Gløshaugen
3129       G012 Gamle elektro Gløshaugen
14894      EL6 Gamle elektro Gløshaugen
3062       EL3 Gamle elektro Gløshaugen
3070       EL2 Gamle elektro Gløshaugen
3078       EL1 Gamle elektro Gløshaugen
3144       G038 Gamle elektro Gløshaugen
3143       G034 Gamle elektro Gløshaugen
3136       G022 Gamle elektro Gløshaugen
14893      EL5 Gamle elektro Gløshaugen

$ ntnu-checkin checkin --room=14894 --from=07:00 --to=18:00
Checked in to EL5 Gamle elektro Gløshaugen from 07:00 to 18:00: OK

$ ntnu-checkin get
CHECKIN-ID                START              END                ROOM-ID      LOCATION
6035fec4c2a0412c2336fe92  Wed Feb 24 07:00   Wed Feb 24 18:00   14894        EL6 Gamle elektro Gløshaugen

$ ntnu-checkin delete 6035fec4c2a0412c2336fe92
Deleted checkin 6035fec4c2a0412c2336fe92: Checkin deleted.

$ ntnu-checkin
Usage of ntnu-checkin:
  ntnu-checkin checkin --room=<room-id> --from=07:00 --to=23:00
    to checkin
  ntnu-checkin search [query]
    to search for rooms
  ntnu-checkin get
    to list checkins
  ntnu-checkin delete [chckin-id]
    to delete checkins
```

## TODO
- (Better) Error handling
- Don't auth each time (works for now)
- Don't auth before printing usage
