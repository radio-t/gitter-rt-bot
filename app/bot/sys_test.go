package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"	
)

func TestSys_OnMessage(t *testing.T) {
	bot := NewSys("./../../data")

	assert.Equal(t, Response{Text: "_никто не знает. пока не надоест_", Send: true}, bot.OnMessage(Message{Text: "доколе?"}))
	assert.Equal(t, Response{Text: "_понг_", Send: true}, bot.OnMessage(Message{Text: "пинг"}))
	assert.Equal(t, Response{Text: "_pong_", Send: true}, bot.OnMessage(Message{Text: "ping"}))

	assert.Equal(t, "/say, say! _– набраться мудрости_\n" +
					"/ping, ping _– ответит pong_\n" +
					"/пинг, пинг _– ответит понг_\n" +
					"/who, /кто, кто?, who? _– ведущие Радио-Т_\n" +
					"/when, /когда, когда?, when? _– расписание эфиров Радио-Т_\n" +
					"/how, /как, как?, how? _– online вещание Радио-Т_\n" +
					"/доколе, доколе? _– день закрытия Радио-Т_\n" +
					"/rules, /правила, rules?, правила? _– правила общения в чате_\n", 
					bot.Help())

}