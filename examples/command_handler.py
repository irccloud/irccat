#!/usr/bin/python
import os
import sys
import re
import subprocess

# If this script is set as your command handler, when any ?command is run in IRC it will
# look in the path defined below for a script matching that command name and run it.
#
# e.g. ?uptime would look in "/usr/share/irccat/" (the default) for any script called
# "uptime", with any extension. It would happily run both uptime.sh and uptime.py, or
# a script in whatever language you like. Command names are limited to [0-9a-z].

path = '/usr/share/irccat/'

# Example of retrieving all the environment variables.
# We only need command here as all the others will be available in the script's environment.
nick = os.environ.get('IRCCAT_NICK')
user = os.environ.get('IRCCAT_USER')
host = os.environ.get('IRCCAT_HOST')
channel = os.environ.get('IRCCAT_CHANNEL')
respond_to = os.environ.get('IRCCAT_RESPOND_TO')
command = os.environ.get('IRCCAT_COMMAND')
args = os.environ.get('IRCCAT_ARGS')

found = False
if re.match('^[a-z0-9]+$', command):
    for filename in os.listdir(path):

        if re.match('^%s\.[a-z]+$' % command, filename):
            found = True

            proc = subprocess.Popen(os.path.join(path, filename), stdout=subprocess.PIPE)
            stdout = proc.stdout

            while True:
                # We do this to avoid buffering from the subprocess stdout
                print(os.read(stdout.fileno(), 65536))
                sys.stdout.flush()

                if proc.poll() is not None:
                    break

if not found:
    print("Unknown command '%s'" % command)
