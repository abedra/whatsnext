# whatsnext

This is a small program that looks at your github repositories and
collects all open issues. If you're like me and have way too many
projects this is a really easy way to quickly see what needs attention.

I have spent absolutely no time on making this well factored or
tested. It was simply a quick and dirty hack to figure out things that
I may have let languish.

## Installation

This project is written in go. You will need to have a working go
environment in order to install it. You can build from source using

```sh
go dep ensure
go build
```

Or via the following command:

```
$ go get github.com/abedra/whatsnext
```

## Setup

You will need to supply a yaml file to the program in order for it to
work correctly. The following example shows what is necessary:

```yaml
users:
  - abedra
orgs:
  - repsheet
token: thisisnotatoken
```

In the users section you can provide any number of users or
organizations. The program will look for issues in any supplied
user. The token is for a github access token. Make sure the token you
generate can access the repositories you would like to examine. The
same rules follow for the orgs section. These are split due to
differing API calls to get the information.

## Contributing

Pull requests always welcome.


