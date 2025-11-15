<div align='center'>
<br />
<img src='./images/logo.png' alt='go-reddit logo' height='150'>

---

<div id='badges' align='center'>

[![Actions Status](https://github.com/jehannes/go-reddit/workflows/tests/badge.svg)](https://github.com/jehannes/go-reddit/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jehannes/go-reddit)](https://goreportcard.com/report/github.com/jehannes/go-reddit)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/jehannes/go-reddit/reddit)](https://pkg.go.dev/github.com/jehannes/go-reddit/reddit)

</div>

</div>

## Overview

**Featured in issues [327](https://golangweekly.com/issues/327) and [347](https://golangweekly.com/issues/347) of Golang Weekly 🎉**

go-reddit is a Go client library for accessing the Reddit API. this is my personal fork with some modifications for a personal project.

You can view Reddit's official API documentation [here](https://www.reddit.com/dev/api/).

## Differences from upstream

**This project has been forked from [sadzeih/go-reddit](https://github.com/sadzeih/go-reddit), who forked it from [vartanbeno/go-reddit](https://github.com/vartanbeno/go-reddit).**

Here are the differences from the upstream:

* Fixing a couple of bugs
* Added gallery post support
* Added media metadata 
* fixed some tests 

## Install

To get a specific version from the list of [versions](https://github.com/jehannes/go-reddit/releases):

```sh
go get github.com/jehannes/go-reddit@vX.Y.Z
```

Or for the latest version:

```sh
go get github.com/jehannes/go-reddit
```

The repository structure for managing multiple major versions follows the one outlined [here](https://github.com/go-modules-by-example/index/tree/master/016_major_version_repo_strategy#major-branch-strategy).

More examples are available in the [examples](examples) folder.

## Contributing

I do not accept contributions at this time. This is primarily for my own use, and am unlikely to be reviewing pull requests or issues.

If you feel the contents of my patches are useful, feel free to fork the repository and maintain your own version. 
Or become the one who takes up Vertanbeno's mantle and unify the umpteen forks of go-reddit out there.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
