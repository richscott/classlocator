# classlocator
A small program to scan through your local Maven repository hierarchy, extract
a list of all the classes in each jar, and write a record of the class and
its containing jarfile path to a SQLite database, for fast, easy searching
later. The root of the examined jarfile hierarchy is `$HOME/.m2/repository`.

## Building
```
$ go build
```

## Running
It will print a line for each jarfile it processes. The SQLite database is
written to `jars.db`. On a Macbook Pro M1, this has been observed to process
4672 jarfiles of various sizes in ~8 seconds.

```
$ ./classlocator
[....]
$ sqlite3 jars.db
SQLite version 3.43.2 2023-10-10 13:08:14
Enter ".help" for usage hints.
sqlite> select * from jarclasses where classname like 'org/apache/spark%/SparkSubmit.class' order by jarfile asc;
org/apache/spark/deploy/SparkSubmit.class|/Users/richscott/.m2/repository/io/armadaproject/armada/armada-cluster-manager_2.13/1.0.0-SNAPSHOT/armada-cluster-manager_2.13-1.0.0-SNAPSHOT-all.jar
org/apache/spark/deploy/SparkSubmit.class|/Users/richscott/.m2/repository/io/armadaproject/armada/armada-cluster-manager_2.13/1.0.0-SNAPSHOT/armada-cluster-manager_2.13-1.0.0-SNAPSHOT.jar
org/apache/spark/deploy/SparkSubmit.class|/Users/richscott/.m2/repository/org/apache/spark/spark-core_2.12/3.5.5/spark-core_2.12-3.5.5.jar
org/apache/spark/deploy/SparkSubmit.class|/Users/richscott/.m2/repository/org/apache/spark/spark-core_2.13/3.3.4/spark-core_2.13-3.3.4.jar
org/apache/spark/deploy/SparkSubmit.class|/Users/richscott/.m2/repository/org/apache/spark/spark-core_2.13/3.5.5/spark-core_2.13-3.5.5.jar
sqlite> .quit
$
```

