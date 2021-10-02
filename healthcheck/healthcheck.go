// Command healthcheck reads a file containing a UNIX UTC timestamp and exits
// nonzero after printing a message to stderr if the allotted time has elapsed
// since a value was last recorded. Config value irc.health_period specifies
// the max allotted time plus a small cushion for lag. It defaults to fifteen
// minutes, which is go-ircevent's default PING frequency (as of
// github.com/thoj/go-ircevent@v0.0.0-20210723090443-73e444401d64). Optional
// config item irc.health_file designates the path to the timestamp file. If
// unset, the program says so (to stderr) and exits zero.
package main

import (
	"log"
	"time"

	"github.com/irccloud/irccat/util"
	"github.com/spf13/viper"
)

var lagInterval = 10 * time.Second // for testing
var defaultPeriod = 15 * time.Minute

func main() {
	viper.SetConfigName("irccat")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	viper.SetDefault("irc.health_period", defaultPeriod)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error reading config file. Exiting.")
	}

	healthFile := viper.GetString("irc.health_file")
	if healthFile == "" {
		log.Println("Config option irc.health_file unset; exiting.")
		return
	}

	freq := lagInterval + viper.GetDuration("irc.health_period")

	if err := util.CheckTimestamp(healthFile, freq); err != nil {
		log.Fatalln(err.Error())
	}
}
