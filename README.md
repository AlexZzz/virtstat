# virtstat
report statistics for libvirt domains

#### It only reports block devices stats for now.

One argument required - domain name or uuid:
```
~# ./virtstat -d sdb instance-0000ef26
2018-10-25 16:49:46
Device:       r/s         w/s     flush/s       rkB/s       wkB/s     r_await     w_await flush_await       err/s
sdb           0           0           0           0           0        0.00        0.00        0.00           0

2018-10-25 16:49:47
Device:       r/s         w/s     flush/s       rkB/s       wkB/s     r_await     w_await flush_await       err/s
sdb           0          35          35           0         140        0.00        0.11       28.05           0

2018-10-25 16:49:48
Device:       r/s         w/s     flush/s       rkB/s       wkB/s     r_await     w_await flush_await       err/s
sdb           0           3           3           0          12        0.00        0.12      137.23           0

^C
```

#### Use `-h` or `--help` for options

