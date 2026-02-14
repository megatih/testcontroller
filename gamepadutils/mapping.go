package gamepadutils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

// mappingParts represents the parsed components of a gamepad mapping string.
type mappingParts struct {
	guid   string
	name   string
	keys   []string
	values []string
}

func addMappingHalfAxisValue(parts *mappingParts, key, value string, sign byte) {
	newKey := string(sign) + key

	valueSuffix := sign
	if len(value) > 0 && value[len(value)-1] == '~' {
		// Invert the sign of the bound axis
		if sign == '+' {
			valueSuffix = '-'
		} else {
			valueSuffix = '+'
		}
	}

	newValue := string(valueSuffix) + value
	// Strip trailing '~'
	if len(newValue) > 0 && newValue[len(newValue)-1] == '~' {
		newValue = newValue[:len(newValue)-1]
	}

	parts.keys = append(parts.keys, newKey)
	parts.values = append(parts.values, newValue)
}

func addMappingKeyValue(parts *mappingParts, key, value string) {
	// Split axis values for easy binding purposes
	for axis := sdl.GamepadAxis(0); axis < sdl.GAMEPAD_AXIS_LEFT_TRIGGER; axis++ {
		if key == axis.GamepadStringForAxis() {
			addMappingHalfAxisValue(parts, key, value, '-')
			addMappingHalfAxisValue(parts, key, value, '+')
			return
		}
	}

	parts.keys = append(parts.keys, key)
	parts.values = append(parts.values, value)
}

func splitMapping(mapping string) mappingParts {
	var parts mappingParts

	if mapping == "" {
		return parts
	}

	// Get the guid
	idx := strings.IndexByte(mapping, ',')
	if idx < 0 {
		parts.guid = mapping
		return parts
	}
	parts.guid = mapping[:idx]
	mapping = mapping[idx+1:]

	// Get the name
	idx = strings.IndexByte(mapping, ',')
	if idx < 0 {
		parts.name = mapping
		return parts
	}
	name := mapping[:idx]
	if len(name) > 0 && name[0] != '*' {
		parts.name = name
	}
	mapping = mapping[idx+1:]

	// Get key:value pairs
	for len(mapping) > 0 {
		colon := strings.IndexByte(mapping, ':')
		if colon < 0 {
			break
		}

		key := mapping[:colon]
		rest := mapping[colon+1:]

		comma := strings.IndexByte(rest, ',')
		var value string
		if comma >= 0 {
			value = rest[:comma]
			mapping = rest[comma+1:]
		} else {
			value = rest
			mapping = ""
		}

		addMappingKeyValue(&parts, key, value)
	}

	return parts
}

func findMappingKey(parts *mappingParts, key string) int {
	for i, k := range parts.keys {
		if k == key {
			return i
		}
	}
	return -1
}

func removeMappingValueAt(parts *mappingParts, index int) {
	parts.keys = append(parts.keys[:index], parts.keys[index+1:]...)
	parts.values = append(parts.values[:index], parts.values[index+1:]...)
}

func convertBAXYMapping(parts *mappingParts) {
	baxyMapping := false
	for i, key := range parts.keys {
		if key == "hint" && parts.values[i] == "SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1" {
			baxyMapping = true
			break
		}
	}

	if !baxyMapping {
		return
	}

	for i, key := range parts.keys {
		switch key {
		case "a":
			parts.keys[i] = "b"
		case "b":
			parts.keys[i] = "a"
		case "x":
			parts.keys[i] = "y"
		case "y":
			parts.keys[i] = "x"
		case "hint":
			if parts.values[i] == "SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1" {
				parts.values[i] = "!SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1"
			}
		}
	}
}

func combineMappingAxes(parts *mappingParts) {
	for i := 0; i < len(parts.keys); i++ {
		key := parts.keys[i]
		if len(key) == 0 || (key[0] != '-' && key[0] != '+') {
			continue
		}

		baseKey := key[1:]
		for axis := sdl.GamepadAxis(0); axis < sdl.GAMEPAD_AXIS_LEFT_TRIGGER; axis++ {
			if baseKey != axis.GamepadStringForAxis() {
				continue
			}

			// Look for matching axis with opposite sign
			var oppositeSign byte
			if key[0] == '-' {
				oppositeSign = '+'
			} else {
				oppositeSign = '-'
			}
			matchingKey := string(oppositeSign) + baseKey
			matching := findMappingKey(parts, matchingKey)
			if matching < 0 {
				break
			}

			currentValue := parts.values[i]
			matchingValue := parts.values[matching]

			if len(currentValue) > 0 && len(matchingValue) > 0 &&
				((currentValue[0] == '-' && matchingValue[0] == '+') ||
					(currentValue[0] == '+' && matchingValue[0] == '-')) &&
				currentValue[1:] == matchingValue[1:] {
				// Combine these axes
				if key[0] == currentValue[0] {
					// Signs match - just strip the sign
					parts.values[i] = currentValue[1:]
				} else {
					// Signs don't match - invert
					parts.values[i] = currentValue[1:] + "~"
				}
				parts.keys[i] = baseKey
				removeMappingValueAt(parts, matching)
				if matching < i {
					i--
				}
			}
			break
		}
	}
}

func joinMapping(parts *mappingParts) string {
	convertBAXYMapping(parts)
	combineMappingAxes(parts)

	guid := parts.guid
	if guid == "" {
		guid = "*"
	}

	name := parts.name
	if name == "" {
		name = "*"
	}

	// Sort keys
	type sortEntry struct {
		key   string
		value string
	}
	entries := make([]sortEntry, len(parts.keys))
	for i := range parts.keys {
		entries[i] = sortEntry{parts.keys[i], parts.values[i]}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	// Move special keys to front/back
	moveSortedEntry := func(key string, front bool) {
		for i, e := range entries {
			if e.key == key {
				entry := entries[i]
				entries = append(entries[:i], entries[i+1:]...)
				if front {
					entries = append([]sortEntry{entry}, entries...)
				} else {
					entries = append(entries, entry)
				}
				break
			}
		}
	}

	moveSortedEntry("face", true)
	moveSortedEntry("type", true)
	moveSortedEntry("platform", true)
	moveSortedEntry("crc", true)
	moveSortedEntry("sdk>=", false)
	moveSortedEntry("sdk<=", false)
	moveSortedEntry("hint", false)

	var b strings.Builder
	b.WriteString(guid)
	b.WriteByte(',')
	b.WriteString(name)
	b.WriteByte(',')
	for _, e := range entries {
		b.WriteString(e.key)
		b.WriteByte(':')
		b.WriteString(e.value)
		b.WriteByte(',')
	}
	return b.String()
}

func recreateMapping(parts *mappingParts) string {
	return joinMapping(parts)
}

func getLegacyKey(key string, baxy bool) string {
	if key == sdl.GAMEPAD_BUTTON_SOUTH.GamepadStringForButton() {
		if baxy {
			return "b"
		}
		return "a"
	}
	if key == sdl.GAMEPAD_BUTTON_EAST.GamepadStringForButton() {
		if baxy {
			return "a"
		}
		return "b"
	}
	if key == sdl.GAMEPAD_BUTTON_WEST.GamepadStringForButton() {
		if baxy {
			return "y"
		}
		return "x"
	}
	if key == sdl.GAMEPAD_BUTTON_NORTH.GamepadStringForButton() {
		if baxy {
			return "x"
		}
		return "y"
	}
	return key
}

func mappingHasKey(mapping, key string) bool {
	parts := splitMapping(mapping)
	i := findMappingKey(&parts, key)
	if i < 0 {
		baxy := strings.Contains(mapping, ",hint:SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1")
		i = findMappingKey(&parts, getLegacyKey(key, baxy))
	}
	return i >= 0
}

func getMappingValue(mapping, key string) string {
	parts := splitMapping(mapping)
	i := findMappingKey(&parts, key)
	if i < 0 {
		baxy := strings.Contains(mapping, ",hint:SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1")
		i = findMappingKey(&parts, getLegacyKey(key, baxy))
	}
	if i >= 0 {
		return parts.values[i]
	}
	return ""
}

func setMappingValue(mapping, key, value string) string {
	if key == "" {
		return mapping
	}

	parts := splitMapping(mapping)
	i := findMappingKey(&parts, key)
	if i >= 0 {
		parts.values[i] = value
	} else {
		parts.keys = append(parts.keys, key)
		parts.values = append(parts.values, value)
	}
	return recreateMapping(&parts)
}

func removeMappingValue(mapping, key string) string {
	parts := splitMapping(mapping)
	i := findMappingKey(&parts, key)
	if i >= 0 {
		removeMappingValueAt(&parts, i)
	}
	return recreateMapping(&parts)
}

// GetElementKey returns the mapping key for a gamepad element.
func GetElementKey(element int) string {
	if element < int(sdl.GAMEPAD_BUTTON_COUNT) {
		return sdl.GamepadButton(element).GamepadStringForButton()
	}

	switch element {
	case SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE:
		return fmt.Sprintf("-%s", sdl.GAMEPAD_AXIS_LEFTX.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE:
		return fmt.Sprintf("+%s", sdl.GAMEPAD_AXIS_LEFTX.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE:
		return fmt.Sprintf("-%s", sdl.GAMEPAD_AXIS_LEFTY.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE:
		return fmt.Sprintf("+%s", sdl.GAMEPAD_AXIS_LEFTY.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE:
		return fmt.Sprintf("-%s", sdl.GAMEPAD_AXIS_RIGHTX.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE:
		return fmt.Sprintf("+%s", sdl.GAMEPAD_AXIS_RIGHTX.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE:
		return fmt.Sprintf("-%s", sdl.GAMEPAD_AXIS_RIGHTY.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE:
		return fmt.Sprintf("+%s", sdl.GAMEPAD_AXIS_RIGHTY.GamepadStringForAxis())
	case SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER:
		return sdl.GAMEPAD_AXIS_LEFT_TRIGGER.GamepadStringForAxis()
	case SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER:
		return sdl.GAMEPAD_AXIS_RIGHT_TRIGGER.GamepadStringForAxis()
	default:
		return ""
	}
}

// MappingHasBindings returns whether a mapping has any button or axis bindings.
func MappingHasBindings(mapping string) bool {
	if mapping == "" {
		return false
	}

	parts := splitMapping(mapping)
	for i := 0; i < int(sdl.GAMEPAD_BUTTON_COUNT); i++ {
		if findMappingKey(&parts, sdl.GamepadButton(i).GamepadStringForButton()) >= 0 {
			return true
		}
	}
	for i := 0; i < int(sdl.GAMEPAD_AXIS_COUNT); i++ {
		if findMappingKey(&parts, sdl.GamepadAxis(i).GamepadStringForAxis()) >= 0 {
			return true
		}
	}
	return false
}

// MappingHasName returns true if the mapping has a controller name.
func MappingHasName(mapping string) bool {
	parts := splitMapping(mapping)
	return parts.name != ""
}

// GetMappingName returns the name from a mapping.
func GetMappingName(mapping string) string {
	parts := splitMapping(mapping)
	return parts.name
}

// SetMappingName sets the name in a mapping, returning a new mapping string.
func SetMappingName(mapping, name string) string {
	if name == "" {
		return mapping
	}

	// Remove leading whitespace
	name = strings.TrimLeft(name, " \t\n\r")

	// Remove commas (field separators)
	name = strings.ReplaceAll(name, ",", "")

	// Remove trailing whitespace
	name = strings.TrimRight(name, " \t\n\r")

	if name == "" {
		return mapping
	}

	parts := splitMapping(mapping)
	parts.name = name
	return recreateMapping(&parts)
}

// GetMappingType returns the gamepad type from a mapping.
func GetMappingType(mapping string) sdl.GamepadType {
	value := getMappingValue(mapping, "type")
	if value == "" {
		return sdl.GAMEPAD_TYPE_UNKNOWN
	}
	return sdl.GetGamepadTypeFromString(value)
}

// SetMappingType sets the type in a mapping, returning a new mapping string.
func SetMappingType(mapping string, typ sdl.GamepadType) string {
	typeStr := typ.GamepadStringForType()
	if typeStr == "" || typ == sdl.GAMEPAD_TYPE_UNKNOWN {
		return removeMappingValue(mapping, "type")
	}
	return setMappingValue(mapping, "type", typeStr)
}

// MappingHasElement returns true if a mapping has the given element bound.
func MappingHasElement(mapping string, element int) bool {
	key := GetElementKey(element)
	if key == "" {
		return false
	}
	return mappingHasKey(mapping, key)
}

// GetElementBinding returns the binding string for an element, or "" if not bound.
func GetElementBinding(mapping string, element int) string {
	key := GetElementKey(element)
	if key == "" {
		return ""
	}
	return getMappingValue(mapping, key)
}

// SetElementBinding sets or clears the binding for an element.
func SetElementBinding(mapping string, element int, binding string) string {
	key := GetElementKey(element)
	if key == "" {
		return mapping
	}
	if binding != "" {
		return setMappingValue(mapping, key, binding)
	}
	return removeMappingValue(mapping, key)
}

// GetElementForBinding returns the element bound to a given binding string.
func GetElementForBinding(mapping, binding string) int {
	if binding == "" {
		return SDL_GAMEPAD_ELEMENT_INVALID
	}

	parts := splitMapping(mapping)
	for i, v := range parts.values {
		if binding == v {
			for element := 0; element < SDL_GAMEPAD_ELEMENT_MAX; element++ {
				key := GetElementKey(element)
				if key != "" && key == parts.keys[i] {
					return element
				}
			}
			break
		}
	}
	return SDL_GAMEPAD_ELEMENT_INVALID
}

// MappingHasBinding returns true if a mapping contains the given binding string.
func MappingHasBinding(mapping, binding string) bool {
	if binding == "" {
		return false
	}

	parts := splitMapping(mapping)
	for _, v := range parts.values {
		if binding == v {
			return true
		}
	}
	return false
}

// ClearMappingBinding removes all elements that use the given binding.
func ClearMappingBinding(mapping, binding string) string {
	if binding == "" {
		return mapping
	}

	parts := splitMapping(mapping)
	modified := false
	for i := len(parts.values) - 1; i >= 0; i-- {
		if binding == parts.values[i] {
			removeMappingValueAt(&parts, i)
			modified = true
		}
	}
	if modified {
		return recreateMapping(&parts)
	}
	return mapping
}

// GetBindingString searches a mapping for a label and returns the binding value.
func GetBindingString(label, mapping string) (string, bool) {
	if mapping == "" {
		return "", false
	}

	var results []string
	idx := strings.Index(mapping, label)
	for idx >= 0 {
		value := mapping[idx+len(label):]
		end := strings.IndexByte(value, ',')
		if end >= 0 {
			value = value[:end]
		}
		results = append(results, value)
		remaining := mapping[idx+len(label):]
		nextIdx := strings.Index(remaining, label)
		if nextIdx < 0 {
			break
		}
		idx = idx + len(label) + nextIdx
	}

	if len(results) == 0 {
		return "", false
	}
	return strings.Join(results, ","), true
}

// GetButtonBindingString returns the binding string for a button from a mapping.
func GetButtonBindingString(button sdl.GamepadButton, mapping string) (string, bool) {
	if mapping == "" {
		return "", false
	}

	label := fmt.Sprintf(",%s:", button.GamepadStringForButton())
	if text, ok := GetBindingString(label, mapping); ok {
		return text, true
	}

	// Try legacy button names
	baxy := strings.Contains(mapping, ",hint:SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1")
	switch button {
	case sdl.GAMEPAD_BUTTON_SOUTH:
		if baxy {
			return GetBindingString(",b:", mapping)
		}
		return GetBindingString(",a:", mapping)
	case sdl.GAMEPAD_BUTTON_EAST:
		if baxy {
			return GetBindingString(",a:", mapping)
		}
		return GetBindingString(",b:", mapping)
	case sdl.GAMEPAD_BUTTON_WEST:
		if baxy {
			return GetBindingString(",y:", mapping)
		}
		return GetBindingString(",x:", mapping)
	case sdl.GAMEPAD_BUTTON_NORTH:
		if baxy {
			return GetBindingString(",x:", mapping)
		}
		return GetBindingString(",y:", mapping)
	}
	return "", false
}

// GetAxisBindingString returns the binding string for an axis direction from a mapping.
func GetAxisBindingString(axis sdl.GamepadAxis, direction int, mapping string) (string, bool) {
	if mapping == "" {
		return "", false
	}

	// Check for explicit half-axis
	var label string
	if direction < 0 {
		label = fmt.Sprintf(",-%s:", axis.GamepadStringForAxis())
	} else {
		label = fmt.Sprintf(",+%s:", axis.GamepadStringForAxis())
	}
	if text, ok := GetBindingString(label, mapping); ok {
		return text, true
	}

	// Get the binding for the whole axis and split if necessary
	label = fmt.Sprintf(",%s:", axis.GamepadStringForAxis())
	text, ok := GetBindingString(label, mapping)
	if !ok {
		return "", false
	}

	if axis != sdl.GAMEPAD_AXIS_LEFT_TRIGGER && axis != sdl.GAMEPAD_AXIS_RIGHT_TRIGGER {
		if len(text) > 0 && text[0] == 'a' {
			// Split the axis
			if len(text) > 0 && text[len(text)-1] == '~' {
				direction *= -1
				text = text[:len(text)-1]
			}
			if direction > 0 {
				text = "+" + text
			} else {
				text = "-" + text
			}
		}
	}
	return text, true
}
