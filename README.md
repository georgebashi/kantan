# 簡単 Kantan

Hi. Kantan deploys [12-factor apps](http://www.12factor.net/), doing all the crap usually needed to get your code into production.  
You might call it a single-host [Heroku](http://www.heroku.com/).

## Deploying your app
Easy:

1. Download the binary and run it:  
<code>wget http://no.releases.yet/sorry && ./kantan</code>
2. Add it as a remote to your 12-factor app's repo:  
<code>git remote add kantan http://localhost:9090/projects/test/repo</code>
3. Push to it:  
<code>git push kantan master</code>

Done. Kantan will build, run, and manage your site.

## Wait, what?
Kantan can be broken down like this:

* You deploy stuff with a single git push, just like [Heroku](http://www.heroku.com/).
* It builds your code with [Heroku's Buildpacks](https://devcenter.heroku.com/articles/buildpacks), like [Mason](https://github.com/ddollar/mason).
* It runs services like [Foreman](https://github.com/ddollar/foreman),
* proxies to them, like [Nginx](http://wiki.nginx.org/Main),
* and watches them, like [Monit](http://mmonit.com/monit/).

Which means you don't have to deploy any of those.

Kantan is a single binary, with zero dependencies, and zero config. You put it in a directory, run it, and push your code to it.  
Kantan takes care of the rest.
