[![Docker Build Status](https://img.shields.io/docker/build/xperimental/ipromnb.svg?style=flat-square)](https://hub.docker.com/r/xperimental/ipromnb/)

# ipromnb

Kernel for [Jupyter Notebooks](http://jupyter.org/) which can query [Prometheus](https://prometheus.io/) servers.

Because notebooks can also contain documentation this makes it ideal for things like

- prototyping new queries for future dashboards
- documenting an outage using metrics

The generated files also contain the results of the queries and so are self-contained and can be read without actively querying the Prometheus server that was used to generate them. This is ideal for sharing thoughts or for archival purposes. GitHub also contains a preview renderer, have a look at [the example file](_examples/Test.ipynb).

## Usage

### Starting an instance

For easy start there is a Docker image: [`xperimental/ipromnb`](https://hub.docker.com/r/xperimental/ipromnb/) which is directly runnable.

There's also a `docker-compose.yml` file in this repo, so manually building and running an image should be as easy as (if you have Docker and docker-compose installed):

```bash
git clone https://github.com/xperimental/ipromnb.git
cd ipromnb
docker-compose up --build
```

This will run a jupyter notebooks instance and map it to port 8888. The directory the repository is in will be mounted as a volume in the "work" directory accessible in the Notebook UI.

Check the `_examples` directory for a simple [example notebook](_examples/Test.ipynb).

### Creating your first notebook

This example assumes that a Prometheus server is available using the URL `http://prometheus:9090/`.

Jupyter notebooks contain blocks (so called "cells") that can either be `Code` or `Markdown`. The Markdown is rendered as HTML and can be used for descriptions. The Code cells are rendered by the so-called "kernel" used for the notebook. This project provides a kernel that will interpret the code cells as queries which are to be sent to a Prometheus server.

For this to work the kernel first needs to know which server to send the queries to, so create a first code block and set its contents to:

```plain
@server=http://prometheus:9090
```

Once you run the cell (by using the keyboard shortcut or the small play button to its left) you should get an output similar to the one below (the time will be different):

```plain
Server: http://prometheus:9090
  Time: 2018-08-07T20:32:13Z - 2018-08-08T20:32:13Z (24h0m0.000000578s)
```

New create a second code cell and type the following into it:

```plain
up
```

Again, run the cell. You should get a small table with all your "up" metrics, similar to the "Console" inside Prometheus.

Now, let's finally create a graph. Again, create a new cell (or resume the one from the "up" example) and put the following content into it:

```plain
graph(sum by(job) (rate(scrape_samples_scraped[30m])))
```

When you run this cell, the output below should be a graph plotting your sample rates per job. The `graph()` function is not part of the Prometheus query language but instead interpreted by the `ipromnb` kernel to produce the graph output.

There are also commands to change the timeframe used for the queries. Modify the first cell (the one with the `@server` command in it) to show the following content:

```plain
@server=http://prometheus:9090
@end=now
@start=end-12h
```

This sets the timeframe to "from 12 hours ago until now". The times can either be given relative to `now`, `start` or `end` or in RFC3339 format (for example `2018-08-08T12:00:00Z`).

Now let's restart the notebook and watch what happens. To do this select the "Restart & Run All" item from the "Kernel" menu at the top. This restarts the kernel and re-runs all the code blocks. You should now have output that reflects the changed timeframe in all outputs.

### Commands interpreted by kernel

The kernel tries to interpret every code cell as a query to the Prometheus server except if it is a command that is directly executed by the kernel itself.

#### Setting options

All the internal commands modify internal state of the kernel, which means that they have an effect for all subsequent cell executions until the kernel is either restarted or another command is issued.

Currently the following commands are understood by the kernel:

- `@server=` sets the Prometheus server used for queries.
- `@start=` sets the start time of the timerange used by range queries.
- `@end=` sets the end time of the timerange used by range queries. This time is also used for instant queries.

The `@start=` and `@end=` commands accept either a RFC3339 formatted timestamp (for example `2018-08-08T12:00:00Z`) or a time relative to either `now`, `start` or `end`:

|                                                          |                    |
| -------------------------------------------------------: | ------------------ |
|                                Set the end time to "now" | `@end=now`         |
|              Set the start time to "12 hours before end" | `@start=end-12h`   |
|                     Set the start time to "24 hours ago" | `@start=now-24h`   |
| Set the end time to "6 hours and 30 minutes after start" | `@end=start+6h30m` |

More than one command can be provided in a single code cell (one per line).

#### Plotting graphs

In addition to commands which are used for changing the kernel options there is another command which controls whether a query will be executed as an "instant" or "range" query yielding either a table of values at the `end` time or a plot of the values between the `start` and `end` time:

```plain
graph(<query>)
```

It has a single optional parameter which can be used to set the minimum of the Y axis of the plot to zero:

```plain
graph0(<query>)
```

## Features (including planned)

This project is still in a very early stage of development which means that only a subset of the planned features are implemented already and also that some existing features might change in the future. Feedback and suggestions are appreciated.

- [x] Send range and instant queries to Prometheus server
- [x] Graph range queries
- [x] Provide a way to set timerange to fixed and dynamic values
- [ ] Make graphs more interactive (currently pre-rendered images)
- [ ] Possibility to test recording rules which are not on the server yet
- [ ] Make it possible to test alerting rules against past data
