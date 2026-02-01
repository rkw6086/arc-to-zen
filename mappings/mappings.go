package mappings

// MapArcIconToSvg maps Arc icon names to Zen SVG icon paths
func MapArcIconToSvg(arcIcon string) string {
	if arcIcon == "" {
		return defaultSvgIcon
	}

	if svg, ok := arcIconToZenSvg[arcIcon]; ok {
		return svg
	}
	return defaultSvgIcon
}

// MapArcColorToZen maps Arc color names to Zen container colors
func MapArcColorToZen(arcColor string) string {
	if arcColor == "" {
		return defaultColor
	}

	if color, ok := arcColorToZen[arcColor]; ok {
		return color
	}
	return defaultColor
}

const (
	defaultSvgIcon = "chrome://browser/skin/zen-icons/selectable/globe.svg"
	defaultColor   = "gray"
)

var arcIconToZenSvg = map[string]string{
	// Work & Business
	"briefcase": "chrome://browser/skin/zen-icons/selectable/briefcase.svg",
	"office":    "chrome://browser/skin/zen-icons/selectable/briefcase.svg",
	"business":  "chrome://browser/skin/zen-icons/selectable/briefcase.svg",
	"build":     "chrome://browser/skin/zen-icons/selectable/build.svg",
	"construct": "chrome://browser/skin/zen-icons/selectable/construct.svg",
	"card":      "chrome://browser/skin/zen-icons/selectable/card.svg",
	"wallet":    "chrome://browser/skin/zen-icons/selectable/wallet.svg",
	"coins":     "chrome://browser/skin/zen-icons/selectable/coins.svg",
	"money":     "chrome://browser/skin/zen-icons/selectable/coins.svg",
	"dollar":    "chrome://browser/skin/zen-icons/selectable/logo-usd.svg",

	// Communication
	"mail":      "chrome://browser/skin/zen-icons/selectable/mail.svg",
	"email":     "chrome://browser/skin/zen-icons/selectable/mail.svg",
	"call":      "chrome://browser/skin/zen-icons/selectable/call.svg",
	"phone":     "chrome://browser/skin/zen-icons/selectable/call.svg",
	"chat":      "chrome://browser/skin/zen-icons/selectable/chat.svg",
	"message":   "chrome://browser/skin/zen-icons/selectable/chat.svg",
	"megaphone": "chrome://browser/skin/zen-icons/selectable/megaphone.svg",

	// Development & Tech
	"code":             "chrome://browser/skin/zen-icons/selectable/code.svg",
	"terminal":         "chrome://browser/skin/zen-icons/selectable/terminal.svg",
	"bug":              "chrome://browser/skin/zen-icons/selectable/bug.svg",
	"extension-puzzle": "chrome://browser/skin/zen-icons/selectable/extension-puzzle.svg",
	"plugin":           "chrome://browser/skin/zen-icons/selectable/extension-puzzle.svg",
	"flask":            "chrome://browser/skin/zen-icons/selectable/flask.svg",
	"test":             "chrome://browser/skin/zen-icons/selectable/flask.svg",

	// Files & Organization
	"folder":   "chrome://browser/skin/zen-icons/selectable/folder.svg",
	"page":     "chrome://browser/skin/zen-icons/selectable/page.svg",
	"document": "chrome://browser/skin/zen-icons/selectable/page.svg",
	"book":     "chrome://browser/skin/zen-icons/selectable/book.svg",
	"bookmark": "chrome://browser/skin/zen-icons/selectable/bookmark.svg",
	"inbox":    "chrome://browser/skin/zen-icons/selectable/inbox.svg",
	"layers":   "chrome://browser/skin/zen-icons/selectable/layers.svg",

	// Media & Entertainment
	"music":           "chrome://browser/skin/zen-icons/selectable/music.svg",
	"video":           "chrome://browser/skin/zen-icons/selectable/video.svg",
	"image":           "chrome://browser/skin/zen-icons/selectable/image.svg",
	"photo":           "chrome://browser/skin/zen-icons/selectable/image.svg",
	"game-controller": "chrome://browser/skin/zen-icons/selectable/game-controller.svg",
	"game":            "chrome://browser/skin/zen-icons/selectable/game-controller.svg",
	"gaming":          "chrome://browser/skin/zen-icons/selectable/game-controller.svg",
	"volume-high":     "chrome://browser/skin/zen-icons/selectable/volume-high.svg",
	"sound":           "chrome://browser/skin/zen-icons/selectable/volume-high.svg",

	// Food & Drink
	"pizza":     "chrome://browser/skin/zen-icons/selectable/pizza.svg",
	"fast-food": "chrome://browser/skin/zen-icons/selectable/fast-food.svg",
	"cafe":      "chrome://browser/skin/zen-icons/selectable/cafe.svg",
	"coffee":    "chrome://browser/skin/zen-icons/selectable/cafe.svg",
	"ice-cream": "chrome://browser/skin/zen-icons/selectable/ice-cream.svg",
	"cutlery":   "chrome://browser/skin/zen-icons/selectable/cutlery.svg",
	"dining":    "chrome://browser/skin/zen-icons/selectable/cutlery.svg",
	"fish":      "chrome://browser/skin/zen-icons/selectable/fish.svg",
	"egg":       "chrome://browser/skin/zen-icons/selectable/egg.svg",

	// Navigation & Places
	"globe":    "chrome://browser/skin/zen-icons/selectable/globe.svg",
	"globe-1":  "chrome://browser/skin/zen-icons/selectable/globe-1.svg",
	"world":    "chrome://browser/skin/zen-icons/selectable/globe.svg",
	"internet": "chrome://browser/skin/zen-icons/selectable/globe.svg",
	"map":      "chrome://browser/skin/zen-icons/selectable/map.svg",
	"location": "chrome://browser/skin/zen-icons/selectable/location.svg",
	"pin":      "chrome://browser/skin/zen-icons/selectable/location.svg",
	"navigate": "chrome://browser/skin/zen-icons/selectable/navigate.svg",
	"compass":  "chrome://browser/skin/zen-icons/selectable/navigate.svg",
	"airplane": "chrome://browser/skin/zen-icons/selectable/airplane.svg",
	"plane":    "chrome://browser/skin/zen-icons/selectable/airplane.svg",

// Home & Personal
	"heart":    "chrome://browser/skin/zen-icons/selectable/heart.svg",
	"star":     "chrome://browser/skin/zen-icons/selectable/star-1.svg", // Classic 5-pointed star
	"star-1":   "chrome://browser/skin/zen-icons/selectable/star-1.svg",
	"sparkle":  "chrome://browser/skin/zen-icons/selectable/star.svg", // Asterisk/sparkle shape
	"favorite": "chrome://browser/skin/zen-icons/selectable/star-1.svg",
	"people":   "chrome://browser/skin/zen-icons/selectable/people.svg",
	"users":    "chrome://browser/skin/zen-icons/selectable/people.svg",
	"eye":      "chrome://browser/skin/zen-icons/selectable/eye.svg",
	"view":     "chrome://browser/skin/zen-icons/selectable/eye.svg",
	"bed":      "chrome://browser/skin/zen-icons/selectable/bed.svg",
	"sleep":    "chrome://browser/skin/zen-icons/selectable/bed.svg",
	"shirt":    "chrome://browser/skin/zen-icons/selectable/shirt.svg",
	"clothing": "chrome://browser/skin/zen-icons/selectable/shirt.svg",

	// Nature & Weather
	"sun":       "chrome://browser/skin/zen-icons/selectable/sun.svg",
	"moon":      "chrome://browser/skin/zen-icons/selectable/moon.svg",
	"cloud":     "chrome://browser/skin/zen-icons/selectable/cloud.svg",
	"water":     "chrome://browser/skin/zen-icons/selectable/water.svg",
	"leaf":      "chrome://browser/skin/zen-icons/selectable/leaf.svg",
	"plant":     "chrome://browser/skin/zen-icons/selectable/leaf.svg",
	"flame":     "chrome://browser/skin/zen-icons/selectable/flame.svg",
	"fire":      "chrome://browser/skin/zen-icons/selectable/flame.svg",
	"lightning": "chrome://browser/skin/zen-icons/selectable/lightning.svg",
	"bolt":      "chrome://browser/skin/zen-icons/selectable/lightning.svg",

	// Education & Learning
	"school":    "chrome://browser/skin/zen-icons/selectable/school.svg",
	"education": "chrome://browser/skin/zen-icons/selectable/school.svg",
	"brush":     "chrome://browser/skin/zen-icons/selectable/brush.svg",
	"art":       "chrome://browser/skin/zen-icons/selectable/brush.svg",
	"palette":   "chrome://browser/skin/zen-icons/selectable/palette.svg",

	// Sports & Recreation
	"american-football": "chrome://browser/skin/zen-icons/selectable/american-football.svg",
	"football":          "chrome://browser/skin/zen-icons/selectable/american-football.svg",
	"baseball":          "chrome://browser/skin/zen-icons/selectable/baseball.svg",
	"paw":               "chrome://browser/skin/zen-icons/selectable/paw.svg",
	"pet":               "chrome://browser/skin/zen-icons/selectable/paw.svg",

	// Security & Safety
	"lock-closed": "chrome://browser/skin/zen-icons/selectable/lock-closed.svg",
	"lock":        "chrome://browser/skin/zen-icons/selectable/lock-closed.svg",
	"key":         "chrome://browser/skin/zen-icons/selectable/key.svg",
	"warning":     "chrome://browser/skin/zen-icons/selectable/warning.svg",
	"alert":       "chrome://browser/skin/zen-icons/selectable/warning.svg",

	// Science & Space
	"rocket":  "chrome://browser/skin/zen-icons/selectable/rocket.svg",
	"planet":  "chrome://browser/skin/zen-icons/selectable/planet.svg",
	"space":   "chrome://browser/skin/zen-icons/selectable/planet.svg",
	"nuclear": "chrome://browser/skin/zen-icons/selectable/nuclear.svg",

	// Misc
	"bell":        "chrome://browser/skin/zen-icons/selectable/bell.svg",
	"flag":        "chrome://browser/skin/zen-icons/selectable/flag.svg",
	"present":     "chrome://browser/skin/zen-icons/selectable/present.svg",
	"gift":        "chrome://browser/skin/zen-icons/selectable/present.svg",
	"tada":        "chrome://browser/skin/zen-icons/selectable/tada.svg",
	"ticket":      "chrome://browser/skin/zen-icons/selectable/ticket.svg",
	"time":        "chrome://browser/skin/zen-icons/selectable/time.svg",
	"clock":       "chrome://browser/skin/zen-icons/selectable/time.svg",
	"trash":       "chrome://browser/skin/zen-icons/selectable/trash.svg",
	"delete":      "chrome://browser/skin/zen-icons/selectable/trash.svg",
	"basket":      "chrome://browser/skin/zen-icons/selectable/basket.svg",
	"cart":        "chrome://browser/skin/zen-icons/selectable/basket.svg",
	"shopping":    "chrome://browser/skin/zen-icons/selectable/basket.svg",
	"skull":       "chrome://browser/skin/zen-icons/selectable/skull.svg",
	"weight":      "chrome://browser/skin/zen-icons/selectable/weight.svg",
	"fitness":     "chrome://browser/skin/zen-icons/selectable/weight.svg",
	"logo-rss":    "chrome://browser/skin/zen-icons/selectable/logo-rss.svg",
	"rss":         "chrome://browser/skin/zen-icons/selectable/logo-rss.svg",
	"stats-chart": "chrome://browser/skin/zen-icons/selectable/stats-chart.svg",
	"chart":       "chrome://browser/skin/zen-icons/selectable/stats-chart.svg",
	"analytics":   "chrome://browser/skin/zen-icons/selectable/stats-chart.svg",
}

// MapArcIconToContainerIcon maps Arc icon names to Firefox container icons
// Firefox containers use a specific set of icons: fingerprint, briefcase, dollar,
// cart, circle, gift, vacation, food, fruit, pet, tree, chill, fence
func MapArcIconToContainerIcon(arcIcon string) string {
	if arcIcon == "" {
		return defaultContainerIcon
	}

	if icon, ok := arcIconToContainerIcon[arcIcon]; ok {
		return icon
	}
	return defaultContainerIcon
}

const defaultContainerIcon = "briefcase"

var arcIconToContainerIcon = map[string]string{
	// Work & Business
	"briefcase": "briefcase",
	"office":    "briefcase",
	"business":  "briefcase",
	"build":     "briefcase",
	"construct": "briefcase",

	// Money & Shopping
	"card":     "dollar",
	"wallet":   "dollar",
	"coins":    "dollar",
	"money":    "dollar",
	"dollar":   "dollar",
	"basket":   "cart",
	"cart":     "cart",
	"shopping": "cart",

	// Food & Drink
	"pizza":     "food",
	"fast-food": "food",
	"cafe":      "food",
	"coffee":    "food",
	"ice-cream": "food",
	"cutlery":   "food",
	"dining":    "food",
	"fish":      "food",
	"egg":       "food",

	// Nature
	"leaf":  "tree",
	"plant": "tree",
	"sun":   "tree",
	"moon":  "chill",
	"cloud": "chill",
	"water": "chill",

	// Personal & Social
	"heart":    "circle",
	"star":     "circle",
	"favorite": "circle",
	"people":   "fingerprint",
	"users":    "fingerprint",
	"eye":      "fingerprint",

	// Travel & Places
	"globe":    "vacation",
	"globe-1":  "vacation",
	"world":    "vacation",
	"internet": "vacation",
	"map":      "vacation",
	"location": "vacation",
	"pin":      "vacation",
	"navigate": "vacation",
	"compass":  "vacation",
	"airplane": "vacation",
	"plane":    "vacation",

	// Animals & Pets
	"paw": "pet",
	"pet": "pet",

	// Gifts & Celebrations
	"present": "gift",
	"gift":    "gift",
	"tada":    "gift",

	// Security
	"lock-closed": "fingerprint",
	"lock":        "fingerprint",
	"key":         "fingerprint",

	// Fruits (for fruit-related icons)
	"fruit": "fruit",

	// Relaxation
	"bed":   "chill",
	"sleep": "chill",

	// Misc - map to circle as generic
	"bell":      "circle",
	"flag":      "circle",
	"ticket":    "circle",
	"time":      "circle",
	"clock":     "circle",
	"music":     "circle",
	"video":     "circle",
	"image":     "circle",
	"photo":     "circle",
	"game":      "circle",
	"gaming":    "circle",
	"code":      "circle",
	"terminal":  "circle",
	"bug":       "circle",
	"rocket":    "circle",
	"planet":    "circle",
	"space":     "circle",
	"flame":     "circle",
	"fire":      "circle",
	"lightning": "circle",
	"bolt":      "circle",
}

var arcColorToZen = map[string]string{
	// Primary colors
	"blue":      "blue",
	"red":       "red",
	"green":     "green",
	"yellow":    "yellow",
	"orange":    "orange",
	"purple":    "purple",
	"pink":      "pink",
	"cyan":      "turquoise",
	"turquoise": "turquoise",

	// Extended colors
	"gray":  "gray",
	"grey":  "gray",
	"black": "black",
	"white": "white",

	// Specific shades
	"light-blue": "blue",
	"dark-blue":  "blue",
	"sky-blue":   "blue",
	"navy":       "blue",

	"light-green": "green",
	"dark-green":  "green",
	"lime":        "green",

	"light-red": "red",
	"dark-red":  "red",
	"crimson":   "red",

	"light-purple": "purple",
	"dark-purple":  "purple",
	"violet":       "purple",
	"indigo":       "purple",

	"light-orange": "orange",
	"dark-orange":  "orange",

	"light-pink": "pink",
	"dark-pink":  "pink",
	"magenta":    "pink",
}
