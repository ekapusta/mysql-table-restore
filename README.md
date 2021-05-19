# mysql-table-restore

Fast MySQL/MariaDB table restore

Implement of copying transportable tablespaces for [MySQL](https://dev.mysql.com/doc/refman/5.7/en/innodb-table-import.html) and [MariaDB](https://mariadb.com/kb/en/innodb-file-per-table-tablespaces/#copying-transportable-tablespaces)

Use when you need to restore many big tables from partial xtrabackup (prepare with export flag)

## Limitations

- same version on source and destination server

- same table structure on source and destination server

## Build

```shell
go build -tags "osusergo netgo"
```

## Usage

```
  -data-dir string
    	Directory with data (default "/var/lib/mysql/")
  -database string
    	Database
  -host string
    	Host (default "tcp(127.0.0.1:3306)")
  -password string
    	Password
  -target-dir string
    	Directory with backuped files (default "./")
  -username string
    	User (default "root")
  -workers int
    	Number of workers (default 1)
```