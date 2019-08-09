debpackage:
	dpkg-buildpackage -d -b -us -uc -P${spec%%~*}	
