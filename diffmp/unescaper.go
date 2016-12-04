package diffmp

import (
	"strings"
)

// unescaper unescapes selected chars for compatability with JavaScript's
// encodeURI.  In speed critical applications this could be dropped since the
// receiving application will certainly decode these fine.  Note that this
// function is case-sensitive.  Thus "%3F" would not be unescaped.  But this
// is ok because it is only called with the output of HttpUtility.UrlEncode
// which returns lowercase hex.
//
// Example: "%3f" -> "?", "%24" -> "$", etc.
var unescaper = strings.NewReplacer(
	"%21", "!", "%7E", "~", "%27", "'",
	"%28", "(", "%29", ")", "%3B", ";",
	"%2F", "/", "%3F", "?", "%3A", ":",
	"%40", "@", "%26", "&", "%3D", "=",
	"%2B", "+", "%24", "$", "%2C", ",", "%23", "#", "%2A", "*")
