module github.com/fumiama/emozi

go 1.20

require (
	github.com/FloatTech/sqlite v1.6.3
	github.com/fumiama/go-base16384 v1.7.0
	github.com/liuzl/gocc v0.0.0-20231231122217-0372e1059ca5
)

require (
	github.com/FloatTech/ttl v0.0.0-20220715042055-15612be72f5b // indirect
	github.com/adamzy/cedar-go v0.0.0-20170805034717-80a9c64b256d // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/liuzl/cedar-go v0.0.0-20170805034717-80a9c64b256d // indirect
	github.com/liuzl/da v0.0.0-20180704015230-14771aad5b1d // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.3.7 // indirect
	modernc.org/libc v1.21.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.4.0 // indirect
	modernc.org/sqlite v1.20.0 // indirect
)

replace modernc.org/sqlite => github.com/fumiama/sqlite3 v1.20.0-with-win386

replace github.com/remyoudompheng/bigfft => github.com/fumiama/bigfft v0.0.0-20211011143303-6e0bfa3c836b
