Installation for development of **antennaApp**
===========================================

**antennaApp** antennaApp provides a simple application that supports [Textcasting](https://textcasting.org) and Link Blogging.
It is built on my experience build my [Antenna](https://github.com/rsdoiel/antenna) personal news aggregator site and my experimental feed reader
project called [Skimmer](https://github.com/rsdoiel/skimmer). It uses the Skimmer project to create a turn key
Antenna style website and curation tool that can run on common computers running Raspberry Pi OS, macOS and Windows.

Quick install with curl or irm
------------------------------

There is an experimental installer.sh script that can be run with the following command to install latest table release. This may work for macOS, Linux and if you’re using Windows with the Unix subsystem. This would be run from your shell (e.g. Terminal on macOS).

~~~shell
curl https://rsdoiel.github.io/antennaApp/installer.sh | sh
~~~

This will install the programs included in antennaApp in your `$HOME/bin` directory.

If you are running Windows 10 or 11 use the Powershell command below.

~~~ps1
irm https://rsdoiel.github.io/antennaApp/installer.ps1 | iex
~~~

### If your are running macOS or Windows

You may get security warnings if you are using macOS or Windows. See the notes for the specific operating system you're using to fix issues.

- [INSTALL_NOTES_macOS.md](INSTALL_NOTES_macOS.md)
- [INSTALL_NOTES_Windows.md](INSTALL_NOTES_Windows.md)


Installing from source
----------------------

### Required software

- Go &gt;&#x3D; 1.25.0
- CMTools &gt;&#x3D; 0.0.40

### Steps

1. git clone https://github.com/rsdoiel/antennaApp
2. Change directory into the `antennaApp` directory
3. Make to build, test and install

~~~shell
git clone https://github.com/rsdoiel/antennaApp
cd antennaApp
make
make test
make install
~~~

