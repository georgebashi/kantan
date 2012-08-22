# 簡単 Kantan

Hi. Kantan deploys 12-factor apps, doing all the crap usually needed to get your code into production. You might call it a single-host heroku.

## Deploying your app

1. Download the binary (none available yet!) and run it
2. Add it as a git remote to your 12-factor app's repo  
(<code>git remote add kantan http://localhost:9090/projects/test/repo</code>)
3. Push to it <code>git push kantan master</code>

Done. Kantan will build, run, and manage your site.

## Wait, what?
Kantan can be broken down like this:

* You deploy stuff with a single git push, just like heroku.
* It builds your code with heroku's buildpacks, like mason.
* It runs services like foreman,
* proxies to them, like nginx,
* and watches them, like monit.

Which means you don't have to deploy any of those.

Kantan is a single binary, with zero dependencies, and zero config. You put it in a directory, run it, and push your code to it. Kantan takes care of the rest.

