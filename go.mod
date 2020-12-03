module github.com/kazeburo/go-mysqlflags

go 1.15

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/mitchellh/mapstructure v1.4.0
	github.com/percona/go-mysql v0.0.0-20200511222729-cd2547baca36
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shirou/gopsutil v3.20.10+incompatible // indirect
	github.com/stretchr/testify v1.6.1
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec
	golang.org/x/sys v0.0.0-20201126233918-771906719818 // indirect
)

replace github.com/mitchellh/mapstructure => github.com/kazeburo/mapstructure v1.4.1-0.20201203061123-1b85cddd5215
