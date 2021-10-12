# fugit

`fugit` is a time tracking tool that is accessed from the command-line. It's
inspired by [https://www.ledger-cli.org/](`ledger-cli`), except it's for time
tracking and not accounting.

This is also my first Go project. ¯\_(ツ)_/¯


`fugit` works with a simple text file where all activities/tasks are kepts as follows:

```
2021-10-12
    08:00-08:30 Breakfast
    08:30-09:12 Read newspaper
    09:20-11:30 Work in project

2021-10-13
    08:05-08:28 Breakfast
    08:28-10:04 Read a book
    10:10-12:08 
```

`fugit` just read the file (located at `$FUGIT_FILE`) and do some computations such as:

```bash
% fugit -d  # today
OK: /path/to/fugit.txt

Read 6 tasks
From 2021-10-12 09:05:00 +0000 UTC
To   2021-10-12 21:40:00 +0000 UTC

Time spent: 4h17m0s

$ fugit -w  # this week
OK: /path/to/fugit.txt

Read 15 tasks
From 2021-10-10 09:25:00 +0000 UTC
To   2021-10-12 21:40:00 +0000 UTC

Time spent: 8h18m0s

% fugit -m  # this month
OK: /path/to/fugit.txt

Read 35 tasks
From 2021-10-01 10:00:00 +0000 UTC
To   2021-10-12 21:40:00 +0000 UTC

Time spent: 26h28m0s
```

As `$FUGIT_FILE` is a regular text file, searching operations and other
things are already implemented with other existing tools such as `grep`.


