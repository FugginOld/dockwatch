## Updating Dockwatch

If dockwatch is monitoring the same Docker daemon under which the dockwatch container itself is running (i.e. if you 
volume-mounted `/var/run/docker.sock` into the dockwatch container) then it has the ability to update itself.  
If a new version of the `fugginold/dockwatch` image is pushed to the Docker Hub, your dockwatch will pull down the 
new image and restart itself automatically.
