package main

import ()

func (p *PBFT) applyCommandLocked(command []byte) {
	commandStr := string(command)
	parts := splitCommand(commandStr)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "SET":
		if len(parts) != 3 {
			return
		}
		key := parts[1]
		value := parts[2]
		p.StateMachine[key] = value

	case "GET":
		if len(parts) != 2 {
			return
		}
		key := parts[1]
		_ = p.StateMachine[key] 
	case "DELETE":
		if len(parts) != 2 {
			return
		}
		key := parts[1]

		delete(p.StateMachine, key)
	default:
		// Unknown command
	}
	
	p.logPutLocked("Applied command to state machine: "+commandStr, GREEN)
}

func splitCommand(command string) []string {
	var parts []string
	current := ""
	for i := 0; i < len(command); i++ {
		if command[i] == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(command[i])
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
