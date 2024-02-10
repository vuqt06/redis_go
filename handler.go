package main

import "sync"

var SETs = map[string]string{}
var SETsMutex = sync.RWMutex{}

var HSETs = map[string]map[string]string{}
var HSETsMutex = sync.RWMutex{}

// ping is a simple handler that returns "PONG"
func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}
	return Value{typ: "string", str: args[0].bulk}
}

// set is a simple handler that sets a key to a value
func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMutex.Lock()
	SETs[key] = value
	SETsMutex.Unlock()

	return Value{typ: "string", str: "OK"}
}

// get is a simple handler that gets a value from a key
func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk

	SETsMutex.RLock()
	value, ok := SETs[key]
	SETsMutex.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

// hset is a simple handler that sets a field in a hash to another hash
func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	key := args[0].bulk
	field := args[1].bulk
	value := args[2].bulk

	HSETsMutex.Lock()
	if _, ok := HSETs[key]; !ok {
		HSETs[key] = map[string]string{}
	}
	HSETs[key][field] = value
	HSETsMutex.Unlock()

	return Value{typ: "string", str: "OK"}
}

// hget is a simple handler that gets a value from a field in a hash
func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	key := args[0].bulk
	field := args[1].bulk

	HSETsMutex.RLock()
	value, ok := HSETs[key][field]
	HSETsMutex.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"HSET": hset,
	"HGET": hget,
}
