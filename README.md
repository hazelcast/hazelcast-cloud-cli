<h1>Hazelcast Cloud CLI</h1>
<p>
  <img alt="release" src="https://github.com/hazelcast/hazelcast-cloud-cli/workflows/Release/badge.svg" />
  <img alt="version" src="https://img.shields.io/github/v/release/hazelcast/hazelcast-cloud-cli">
  <img alt="go-version" src="https://img.shields.io/github/go-mod/go-version/hazelcast/hazelcast-cloud-cli" />
  <img alt="issues" src="https://img.shields.io/github/issues-raw/hazelcast/hazelcast-cloud-cli">
  <a href="https://github.com/hazelcast/hazelcast/blob/master/LICENSE" target="_blank">
    <img alt="License: Apache License 2.0" src="https://img.shields.io/badge/License-Apache License 2.0-yellow.svg" />
  </a>
</p>

> Hazelcast Cloud CLI is known as `hzcloud` is a command line tool to make actions on Hazelcast Cloud easily. Hazelcast Cloud offers the leading in-memory computing platform, Hazelcast IMDG, as a fully managed service that integrates with your existing virtual private cloud.
 <img alt="Screenshot" src="https://user-images.githubusercontent.com/1237982/97022384-b901a780-155c-11eb-87b4-5f4e945a0f1c.png" />

## Installing `hzcloud`
### Using a Package Manager (Homebrew)
```sh
brew tap hazelcast/hz
brew install hzcloud
```
### Downloading a Release from GitHub
Visit the [Releases page](https://github.com/hazelcast/hazelcast-cloud-cli/releases) for the
[`hzcloud` GitHub project](https://github.com/hazelcast/hazelcast-cloud-cli), and find the version for your operating system and architecture. Then place it into your directory with name `hzcloud` or `hzcloud.exe` for Windows.

**Linux** 
```sh
wget \
  https://github.com/hazelcast/hazelcast-cloud-cli/releases/latest/download/hzcloud-linux-amd64 \
  -O /usr/local/bin/hzcloud && chmod +x /usr/local/bin/hzcloud
```
**Windows** 
```sh
curl -o hzcloud.exe `
  https://github.com/hazelcast/hazelcast-cloud-cli/releases/latest/download/hzcloud-windows-amd64
```
On Windows, in order to use `hzcloud` on everywhere you need to put `hzcloud.exe` into your PATH.

**MacOS** 
```sh
wget \
  https://github.com/hazelcast/hazelcast-cloud-cli/releases/latest/download/hzcloud-darwin-amd64 \
  -O /usr/local/bin/hzcloud && chmod +x /usr/local/bin/hzcloud
```
## Authentication with Hazelcast Cloud
After a successful installation, in order to use, you need to authenticate with Hazelcast Cloud by providing access tokens, which can be created from `Developers` tab in [Hazelcast Cloud](https://cloud.hazelcast.com/settings/developer). You can check how to generate API Key and API Secret following the [Hazelcast Cloud Documentation](https://docs.cloud.hazelcast.com/docs/developer).

### Using Environment Variables (Option 1)
You can pass your API Key as `HZ_CLOUD_API_KEY` and API Secret as `HZ_CLOUD_API_SECRET` on your environment variables. `hzcloud` will use these them to authenticate with Hazelcast Cloud

### Using Login Command (Option 2)
You can use login command to provide your API Key and Secret from `hzcloud`.
```sh
$ hzcloud login
-  Api Key: SAMPLE_API_KEY
-  Api Secret: SAMPLE_API_SECRET
```

## :rocket: Examples
You can use `hzcloud` to interact with resources on **Hazelast Cloud**. You can find some examples to begin with.

- **Create a Starter Cluster**
```sh
hzcloud starter-cluster create \
  --cloud-provider=aws \
  --cluster-type=FREE \
  --name=mycluster \
  --region=us-west-2 \
  --total-memory=0.2 \
  --hazelcast-version=5.0.2
```
Also, you can check other parameters with help command
```sh
hzcloud starter-cluster -help
```

- **List Starter Clusters**
```sh
hzcloud starter-cluster list
```

- **Create a Enterprise Cluster**
```sh
hzcloud enterprise-cluster create \
  --name=mycluster \
  --cloud-provider=aws \
  --region=eu-west-2 \
  --zones=eu-west-2b \
  --hazelcast-version=5.0.2 \
  --instance-type=m5.large \
  --cidr-block=10.0.80.0/16 \
  --native-memory=4 \
  --wait
```
Also, you can check other parameters with help command
```sh
hzcloud enterprise-cluster -help
```

- **List Enterprise Clusters**
```sh
hzcloud enterprise-cluster list
```

- **Update hzcloud**
```sh
hzcloud version update
```

## 🏷️ Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/hazelcast/hazelcast-cloud-cli/tags).

## ⭐️ Built With

* [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
* [Color](https://github.com/fatih/color) - Color package for Go
* [Go-Pretty](https://github.com/jedib0t/go-pretty) - Pretty print tables and more in Go

## 🤝 Contributing

Contributions, issues and feature requests are welcome!<br />Feel free to check [issues page](https://github.com/hazelcast/hazelcast-cloud-cli/issues).

## 📝 License

Copyright © 2020 [Hazelcast](https://github.com/hazelcast).<br />
This project is [Apache License 2.0](https://github.com/hazelcast/hazelcast-cloud-cli) licensed.<br /><br />
<img alt="logo" width="300" src="https://cloud.hazelcast.com/static/images/hz-cloud-logo.svg" />
