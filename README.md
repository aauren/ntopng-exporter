# ntopng-exporter
Is a metric exporter for ntopng which scrapes the API that ntopng exposes and then publishes it as a metric.

# Installing
Go to the GitHub [Releases Page](https://github.com/aauren/ntopng-exporter/releases) and download a release for your system. This project publishes binaries for MacOS (darwin), Linux, Windows in x86, amd64, and arm architectures.

# Configuring
ntopng-exporter tries to setup a few sensible defaults for you, but there are some things that it needs to know in order to run correctly. Specifically, it needs to know the IP and Port of your ntopng setup, what interfaces you want to monitor on the ntopng machine, and the username/password for ntopng (if you have one set).

By default, your release archive will come with a [sample config file](https://github.com/aauren/ntopng-exporter/blob/main/config/ntopng-exporter.yaml) that should outline all of the available configuration options as well as point out defaults where they exist.

Modify the default config in any way that you need, and then copy it to one of the following configuration locations (listed in order of precedence):
* `<user's home directory>/.ntopng-exporter/ntopng-exporter.yaml`
* `/etc/ntopng-exporter/ntopng-exporter.yaml`
* `./config/ntopng-exporter.yaml` (where `./` indicates the working directory that ntopng-exporter is using)

If you configure authentication options for ntopng-exporter, then your config file will contain sensitive information. As such, it is recommended that users change the permissions of the exporter so that it is not widely readable:

For Linux this would be done with:
```
chown <user_ntopng_runs_as> <path_to_config_file>
chmod 700 <path_to_config_file>
```

# Supported Versions of ntopng
Because ntopng's API has shifted quite a bit over the last few versions, this exporter is only able to support version 4.2 and above of ntopng. While I have not specifically tested it against 4.0, I believe that it would not work correctly. Additionally, I can verify that it does not work at all against version 3.8.X or below.

# Running ntopng-exporter
## Linux
If you want to run ntopng-exporter on Linux I recommend copying the [systemd unit file](https://github.com/aauren/ntopng-exporter/blob/main/resources/ntopng-exporter.service) (also included in your download archive in the `/resources` directory) to your local system and having systemd manage it so that it starts when your machine starts.

If you do not run ntopng-exporter on the same host that you run ntopng on, you'll want to modify your exporter and remove: `After=ntopng.service`

To run ntopng-exporter as a service do the following:
```
# From your ntopng unpack directory
sudo cp resources/ntopng-exporter.service /etc/systemd/user
sudo systemctl enable ntopng-exporter
sudo systemctl start ntopng-exporter
systemctl status ntopng-exporter
```

If you run it this way, you'll want to put the configuration file in a system wide path like: `/etc/ntopng-exporter`

### Root Concerns
If you executed the procedure above, systemd will run ntopng-exporter as root which is absolutely not needed and should be a security concern. It is recommended to create a separate user for ntopng-exporter and change the systemd file appropriately.

A basic procedure for this would look like the following:
```
sudo useradd -r ntopngexport
sudo chown ntopngexport /etc/ntopng-exporter/ntopng-exporter.yaml
```

The `-r` flag creates the user as a system user which does not have a password, a home directory, or the ability to login

Then change the systemd unit file as such:
```
[Service]
User=ntopngexport
Group=nobody
```

# Running ntopng-exporter Remotely
ntopng-exporter can be run from any machine, it does not have to run on the host that is running ntopng. It only needs to be able to access the ntopng API endpoint. You can usually check that your API is available using something like the following curl command:
```
curl --cookie "user=admin; password=admin" "http://<ntopng_host>:<ntopng_port>/lua/rest/v1/get/ntopng/interfaces.lua"
```

If that responds, then you should be able to add that endpoint to your .yaml configuration file with confidence.

# Releasing
Releases are generated automatically using [goreleaser](https://goreleaser.com/quick-start/). All that is needed is to [install goreleaser](https://goreleaser.com/install/), [export a repo token from GitHub](https://github.com/settings/tokens/new) (assuming you are releasing to your own fork or that you've been given access to the parent project), and then follow the instructions below:
* Export GitHub Token: `export GITHUB_TOKEN="YOUR_GH_TOKEN"`
* Tag a New Release: `git tag -a v0.1.0 -m "Initial Release"`
* Push Release: `git push origin v0.1.0`
* Run goreleaser: `goreleaser release --rm-dist`
