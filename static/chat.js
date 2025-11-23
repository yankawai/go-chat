// websocket connect
const ws = new WebSocket('ws://localhost:8080/ws')

// dom elements
const colorName = document.getElementById('colorName')
const chatlog = document.getElementById('chatlog') // chatlog
const msgInput = document.getElementById('msgInput') // message user input
const sendBtn = document.getElementById('sendBtn') // user send button [ send]
const nameInput = document.getElementById('nameInput') // user  name input
const chat = document.getElementById('chat') //chat
const connectBtn = document.getElementById('connectBtn') // connect button

let username = null
let color = null

// receive message
ws.onmessage = event => {
	//
	try {
        const payload = JSON.parse(event.data);


        const msgDiv = document.createElement('div');
		msgDiv.classList.add('message');


		const nameSpan = document.createElement('span');
		nameSpan.classList.add('username');
		nameSpan.textContent = payload.user; + ": ";
		nameSpan.style.color = payload.color || "#000";

		msgDiv.appendChild(nameSpan);

		msgDiv.appendChild(document.createTextNode(': ' + payload.msg));
		chatlog.appendChild(msgDiv);
		chatlog.scrollTop = chatlog.scrollHeight;
	} catch (error) {
		console.error('Error processing message:', error)
	}
}

// send message
sendBtn.onclick = () => {
	const message = msgInput.value.trim()
	if (message === '') return
	ws.send(JSON.stringify({
		user: username,
		color: color,
		msg : message,
	}))
	msgInput.value = ''
}

// Choose color button
colorName.onchange = e => {
	color = e.target.value
	if (!color === '')  {
		console.log("")
		return
	}
	console.log('[🟣] Color selected: ',color)
}
// chanel  button
connectBtn.onclick = () => {
	username = nameInput.value.trim()
	if (username === '') return
	chat.style.display = 'block'
	intro.style.display = 'none'
}

// Дополнительно: Enter для отправки
msgInput.addEventListener('keypress', e => {
	if (e.key === 'Enter') sendBtn.click()
})
