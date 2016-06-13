package goredis

type Command struct {
	CommandName string
	Args        []interface{}
}

func NewCommand(commandName string, args ...interface{}) (command Command) {
	command.CommandName = commandName
	command.Args = args
	return command
}

func (this Command) Append(args interface{}) Command {
	this.Args = append(this.Args, args)
	return this
}

type Commands []Command

func (this Commands) Append(command Command) Commands {
	this = append(this, command)
	return this
}
