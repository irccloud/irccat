package dispatcher

import "strings"

var colourMap map[string]string = map[string]string{
	"NORMAL":    "\x0f",
	"BOLD":      "\x02",
	"UNDERLINE": "\x1f",
	"REVERSE":   "\x16",
	"WHITE":     "\x0300",
	"BLACK":     "\x0301",
	"DBLUE":     "\x0302",
	"DGREEN":    "\x0303",
	"RED":       "\x0304",
	"BROWN":     "\x0305",
	"PURPLE":    "\x0306",
	"ORANGE":    "\x0307",
	"YELLOW":    "\x0308",
	"GREEN":     "\x0309",
	"TEAL":      "\x0310",
	"CYAN":      "\x0311",
	"BLUE":      "\x0312",
	"PINK":      "\x0313",
	"DGRAY":     "\x0314",
	"GRAY":      "\x0315",
}

func replaceFormatting(msg string) string {
	for colour, code := range colourMap {
		msg = strings.Replace(msg, "%"+colour, code, -1)
		msg = strings.Replace(msg, "#"+colour, code, -1)
	}
	return msg
}
