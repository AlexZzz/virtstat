# virtstat
report statistics for libvirt domains

#### It only reports block devices stats for now.

One argument required - domain name or uuid:
```
~# ./virtstat instance-0003dab3
2018-10-06 15:56:28
Device:     r/s         w/s       rkB/s       wkB/s
hda           0           0           0           0
vda           0           0           0           0

2018-10-06 15:56:29
Device:     r/s         w/s       rkB/s       wkB/s
hda           0           0           0           0
vda          62        2561        1420      164566

2018-10-06 15:56:30
Device:     r/s         w/s       rkB/s       wkB/s
hda           0           0           0           0
vda          36         765         783       39593

2018-10-06 15:56:31
Device:     r/s         w/s       rkB/s       wkB/s
hda           0           0           0           0
vda          33          16         818          64

^C
```

#### Use `-h` or `--help` for options

