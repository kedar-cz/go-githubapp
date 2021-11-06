# Introduction 

This is a small POC to develop a GitHub Application that will deliver basic value and serve as an infrastructure for future functionality. Following GitHub events have been captured in this POC

* Addition and Deletion of a git branch 
* Push Events 
* Invoking GitHub REST endpoints to capture 
  * GitHub Application Configuration 
  * Fetch all the commits associated with the `main` branch of a repository

# Choice of golang Library 

Following two golang libraries were evaluated for this POC: 
* [go-githubapp](http://godoc.org/github.com/palantir/go-githubapp)
* [go-github-app-boilerplate](https://github.com/sharkySharks/go-github-app-boilerplate)

`go-githubapp` was choosen for this POC development for the following reasons

* Compare to `go-github-app-boilerplate` it has relatively larger community 
* The documentation is better than that of `go-github-app-boilerplate` 
* Support for both `sync` and `async` calls 


# Limitation Of Library 

The `go-githubapp` provides a wrapper around the main [go-github](https://github.com/google/go-github) library. There are certain `go-github` limitations when it comes to building the GitHub Application 

* Support for Webhook deliveries 

# Getting Stated Guide 

# Technical consideration 
* JWT
* Multiple user/repo using same GitHub app 
* Fallback strategy to capture the events when the app is down 
  