# dss-fuse

Locally edit your [Dataiku DSS](https://www.dataiku.com) code recipes
with **any** text editor.

This project aims to be an alternative to the
[vscode](https://github.com/dataiku/dss-integration-vscode)
and [PyCharm](https://github.com/dataiku/dss-integration-pycharm)
plugins for DSS. I started working on this in frustration with the
vscode plugin. I am not sure what the problem is (I never bothered
investigating it further) but the damn plugin stops saving any changes
after a few minutes of using it. I end up copying code from vscode
and pasting it into the super limited online text editor DSS provides.

# Building

As simple as:

```sh
go build
```

# Set up and Usage

- Install `fusermount3`... On Debian/Ubuntu do:

```sh
sudo apt install fuse3
```

- First [create a configuration file for DSS](https://github.com/dataiku/dss-integration-pycharm#configuration)
in the dataiku config directory (~/.dataiku). Note that this
plugin uses the same configuration that PyCharm and vscode uses.
So, if you already have any of those configured, this should just work.

```json
# ~/.dataiku/config.json
{
  "dss_instances": {
    "default": {
      "url": "http(s)://DSS_HOST:DSS_PORT/",
      "api_key": "Your API key secret",
      "no_check_certificate": false
    },
  },
  "default_instance": "default"
}
```

- Create a mount point for the DSS filesystem:

```sh
mkdir ~/dss
```

- Mount the filesystem

```sh
./dss-fuse ~/dss || umount ~/dss
```

- Go ahead and browse the directory

```sh
ls ~/dss
```

# Caveats

- Windows is not supported (works fine in WSL though)
- [Mac is not supported!](https://github.com/bazil/fuse/issues/224)
- No way to run recipes as in PyCharm or vscode. I am planning
on adding a separate binary that will do this.
- Bugs ahoy!!! I am new to
[FUSE](https://www.kernel.org/doc/html/latest/filesystems/fuse.html)
and have very little experience with Go