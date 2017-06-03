#Phone pool generator

Tool to generate phone # sequences out of CSV source file with pool data.

Keeps track of pools used for generation and will not use them again unless 'flush' action performed.
Opens output file for addition, so it is safe to use this tool with same settings after media access error or user interrupt.
Worst to happen is one pool to be generated twice.

Initially created to deal with malformed CSV data created by russian hackers, so supports reformatting to proper CSV.

Requires PostgresSQL UTF-8 encoded database created to store temporarily parsed phone pools

Required CSV data format:

`code,pool_start,pool_end,pool_capacity,pool_owner,pool_region`

Usage example:

```
go get github.com/rcmonitor/phone/pool
go build
./pool -h
cp config/database_example.json config/database.json
nano config/database.json
mkdir log,data
wget -P data https://www.rossvyaz.ru/docs/articles/Kody_DEF-9kh.csv
./pool -a format ./data/Kody_DEF-9kh.csv
./pool -a create
./pool -a parse ./data/Kody_DEF-9kg_formatted.csv
./pool -a generate -b 901
./pool -a flush
./pool -a generate -s 926,925 ./data/megafon.txt

