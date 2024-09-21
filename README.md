<p align="center">
  <picture>
    <source width="50%" media="(prefers-color-scheme: dark)" srcset="ui/assets/img/sourcegraph-logo-dark.svg">
    <img width="50%" alt="Gatherstack logo" src="ui/assets/img/sourcegraph-logo-light.svg">
  </picture>
</p>

# Gatherstack &ndash; a libre, OSS fork of [Sourcegraph](https://sourcegraph.com).

Search all of your repositories across all branches and all code hosts.

This is based on the last OSS, Apache 2.0 commit provided by Sourcegraph Inc., which you can find upstream here: <https://github.com/sourcegraph/sourcegraph-public-snapshot/commit/1cd36d2dbbd2a9ab638cc437d208d2717eaefb0b>.

## Learn more about how to use Gatherstack

You can refer to Sourcegraph's legacy documentation for Sourcegraph 5.1: <https://docs-legacy.sourcegraph.com/@5.1>.
The various 

## How to build

The following steps have been verified to work on Debian 12 (amd64).
They might work on other distributions and architectures as well, although they have not yet been tested.

The following commands need to be run as root (although feel free to adapt if you want to run them as a different user).

```shell
# Install apt dependencies
apt install -y \
  apparmor \
  apparmor-utils \
  build-essential \
  docker.io \
  git \
  parallel \
  ;

# Install specific node version
wget https://nodejs.org/dist/v16.18.1/node-v16.18.1-linux-x64.tar.xz
tar -xJf node-v16.18.1-linux-x64.tar.xz
mv node-v16.18.1-linux-x64 /usr/local/node-v16.18.1
ln -s /usr/local/node-v16.18.1/bin/node /usr/bin/node
ln -s /usr/local/node-v16.18.1/bin/npm /usr/bin/npm

# Install specific golang version
wget https://golang.org/dl/go1.19.8.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.19.8.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install required pnpm version
curl -fsSL https://get.pnpm.io/install.sh | env PNPM_VERSION=8.3.0 sh -

# Check out the source code
git clone https://github.com/pitkley/gatherstack

# Build gatherstack
cd gatherstack/cmd/server/
./pre-build.sh
env IMAGE=pitkley/gatherstack:self-built VERSION=5.1.0 ./build.sh
```

This builds a single Docker image for Gatherstack.
The most basic way to run Gatherstack is like this:

```shell
docker container run \
  --name gatherstack \
  --detach \
  -p 7080:7080 \
  --volume ./gs-config:/etc/sourcegraph \
  --volume ./gs-data:/var/opt/sourcegraph \
  pitkley/gatherstack:self-built
```

After this you can access Gatherstack at <http://localhost:7080>.

## Project goal and naming

This project's goal is to preserve the last Apache 2.0 licensed version of Sourcegraph, and make building and running easy, accessible and self-sustained.

To reduce any confusion with Sourcegraph, this project is named "Gatherstack", although there is no intention to remove every occurrence of "Sourcegraph" from the codebase just for the sake of it.
Instead, all user-facing interfaces should show "Gatherstack", to avoid the users accidentally reaching out to Sourcegraph with issues that Sourcegraph could not support with.

It currently is not a goal to extend Gatherstack with new features.

## Contributing

Currently, no contributions to this repository are accepted.
The risk is that code contributed referenced Sourcegraph's non-OSS enterprise code, making the contribution incompatible with the Apache 2.0 license.

## Affiliation

This project has no affiliation with Sourcegraph Inc. or any of its affiliates.
"Sourcegraph" is a trademark of Sourcegraph Inc.

## License

As of commit `5e0b4eb503df02321c1577f423fa830b4b28eed5` all code in this repository is licensed under the Apache 2.0 license, see [LICENSE](LICENSE) or <https://www.apache.org/licenses/LICENSE-2.0>.

Commits including and prior to commit `1cd36d2dbbd2a9ab638cc437d208d2717eaefb0b` can contain code under certain paths that is licensed under the Sourcegraph "Enterprise License", the details of which you can find in the `README.md` and in the `LICENSE.enterprise` file in the respective commits.

All code up to and including commit `1cd36d2dbbd2a9ab638cc437d208d2717eaefb0b` is copyright 2018-2023 Sourcegraph Inc.
