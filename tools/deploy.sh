#!/usr/bin/env bash
#	deploy.sh

if [[ -z $1 ]]; then
	echo "Usage: ./migrate.sh path/to/revision/dir"
	exit 1
else 
	revdir=$1
	echo "revision directory is $revdir"
fi

echo "run migration"
systemctl stop telepyth.service
cp $revdir/telepyth-srv /usr/bin
cp $revdir/bolt.db /var/lib/telepyth
systemctl start telepyth.service
echo "done."
