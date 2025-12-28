# ntopng-exporter

A metric exporter for ntopng which scrapes the API that ntopng exposes and then publishes it as a metric.

This is an example of the types of Grafana dashboards that can be made from these metrics:
![Grafana Example](/docs/grafana_example.png)

## Current State (as of May of 2024)

At this point, I consider ntopng-exporter more or less feature complete. I am continuing to use it personally, and will
continue to update it's dependencies, go versions, etc. on a quarterly basis or so. If there is anything that you find
is broken feel free to open an issue and I'll take a look at it when I get a chance. However, in terms of features, I'm
pretty limited based upon what is available to me from the ntopng API upstream. Unless they change it significantly, it
is unlikely that many features that most people might want are available.

However, if you do find something that is available via the API that I have so far not included here, PR's are always
welcome!

## Installing

Go to the GitHub [Releases Page](https://github.com/aauren/ntopng-exporter/releases) and download a release for your
system. This project publishes binaries for MacOS (darwin), Linux, Windows in x86, amd64, and arm architectures.

## Configuring

ntopng-exporter tries to setup a few sensible defaults for you, but there are some things that it needs to know in order
to run correctly. Specifically, it needs to know the IP and Port of your ntopng setup, what interfaces you want to
monitor on the ntopng machine, and the username/password for ntopng (if you have one set).

By default, your release archive will come with a [sample config file](https://github.com/aauren/ntopng-exporter/blob/main/config/ntopng-exporter.yaml)
that should outline all of the available configuration options as well as point out defaults where they exist.

Modify the default config in any way that you need, and then copy it to one of the following configuration locations
(listed in order of precedence):

- `<user's home directory>/.ntopng-exporter/ntopng-exporter.yaml`
- `/etc/ntopng-exporter/ntopng-exporter.yaml`
- `./config/ntopng-exporter.yaml` (where `./` indicates the working directory that ntopng-exporter is using)

If you configure authentication options for ntopng-exporter, then your config file will contain sensitive information.
As such, it is recommended that users change the permissions of the config file so that it is not widely readable:

For Linux this would be done with:

```sh
chown <user_ntopng_runs_as> <path_to_config_file>
chmod 700 <path_to_config_file>
```

## Supported Versions of ntopng

Because ntopng's API has shifted quite a bit over the last few versions, this exporter is only able to support version
4.2 and above of ntopng. While I have not specifically tested it against 4.0, I believe that it would not work
correctly. Additionally, I can verify that it does not work at all against version 3.8.X or below.

In ntopng version 5.X they introduced the v2 REST API and deprecated the v1 REST API. While the v2 API appears to be
mostly backwards compatible with the v1 API, it is a different endpoint. After reviewing all of the code changes it
became evident that supporting both versions of the API in a single ntopng-exporter version is not a small task.
Therefore, I am releasing versions of ntopng-exporter that are compatible with either the v1 API or the v2 API and not
both.

For ntopng version 5.X and below, please use ntopng-exporter version `v0.X`

For ntopng version 5.X and above, please use ntopng-exporter version `v1.X`

You'll notice that ntopng version 5.X is compatible with both. For more information please see:
[ntopng release announcment](https://github.com/ntop/ntopng/releases/tag/5.0)

## Running ntopng-exporter

### systemd

If you want to run ntopng-exporter on Linux I recommend copying the [systemd unit file](https://github.com/aauren/ntopng-exporter/blob/main/resources/ntopng-exporter.service)
(also included in your download archive in the `/resources` directory) to your local system and having systemd manage
it so that it starts when your machine starts.

If you do not run ntopng-exporter on the same host that you run ntopng on, you'll want to modify your version of the
unit file and remove: `After=ntopng.service`

To run ntopng-exporter as a service do the following:

```sh
# From your ntopng-exporter unpack directory
sudo cp resources/ntopng-exporter.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable ntopng-exporter
sudo systemctl start ntopng-exporter
systemctl status ntopng-exporter
```

If you run it this way, you'll want to put the configuration file in a system wide path like: `/etc/ntopng-exporter`

Currently the exporter just writes all relevant output, including errors, to stdout. If you have any problems with it
starting or it remaining started, be sure to look through the systemd journal for any errors with something like the
following: `journalctl -u ntopng-exporter --since="5m ago"`

#### Root Concerns

If you executed the procedure above, systemd will run ntopng-exporter as root which is absolutely not needed and may
be a security concern. It is recommended to create a separate user for ntopng-exporter and change the systemd file
appropriately.

A basic procedure for this would look like the following:

```sh
sudo useradd -r ntopngexport
sudo chown ntopngexport /etc/ntopng-exporter/ntopng-exporter.yaml
```

The `-r` flag creates the user as a system user which does not have a password, a home directory, or the ability to
login.

Then change the systemd unit file as such:

```service
[Service]
User=ntopngexport
Group=nobody
```

### docker

Another option is to use docker to deploy the service. You can build the docker image yourself, or use one published
on [Docker Hub](https://hub.docker.com).

To run the docker image, a volume containing the `ntopng-exporter.yaml` configuration file named `/config` must be
mounted and the port needs to be exposed, either with a dynamic port allocation with the `--publish-all` option or
with a specific port mapping with `--publish`. An example docker invocation that runs the latest published
ntopng-exporter would look something like this:

```sh
docker run --volume /some/directory/with/config:/config --publish 3001:3001 aauren/ntopng-exporter
```

To build the image yourself, make sure that you have at least docker 17.05 installed, and build using the command
`docker build .` in the checked out repository.

## Running ntopng-exporter Remotely

ntopng-exporter can be run from any machine, it does not have to run on the host that is running ntopng. It only needs
to be able to access the ntopng API endpoint. You can usually check that your API is available using something like
the following curl command:

```sh
curl --cookie "user=admin; password=admin" "http://<ntopng_host>:<ntopng_port>/lua/rest/v1/get/ntopng/interfaces.lua"
```

If that responds, then you should be able to add that endpoint to your .yaml configuration file with confidence.

## Metrics Emitted

Currently, the only supported metric system is Prometheus, however, the design should be extensible to other metric
systems. There are [examples](/docs/ntopng_exporter_example_metrics.md) of the Prometheus metrics emitted in the docs
section of this repo.

## Grafana Dashboard

There is an example Grafana dashboard that user's can use contained within this repository at [grafana-dashboard.json](/resources/grafana-dashboard.json)

You can also find an example of the dashboard on Grafana's public dashboard service [ntopng-exporter](https://grafana.com/grafana/dashboards/20071)
