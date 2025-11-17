
// generate random username_color
function getRandomColor() { 
	const colors = 
	["#e22","#2df","#45f","#0f0f","#0f0f"];
	return colors[Math.floor(Math.random() * colors.length)];
}

// nickname & color 
const userID = "User-" + Math.floor(Math.random() * 1000)
const userColor =  getRandomColor();

// websocket connect 
const ws = new WebSocket("ws://localhost:8080/ws")
// dom elements 	 
const chatlog = document.getElementById("chatlog")
const msgInput = document.getElementById("msgInput")
const sendBtn = document.getElementById("sendBtn")


//recieve message 
ws.onmessage = () =>  { 
	const msgDiv = document.createElement("div");
	msgDiv.classList.add("message");

	const [sender, ...textParts] = event.data.split(": ");
	const text = textParts.join(": ");

	const nameSpan = document.createElement("span");
	nameSpan.classList.add("username");
	nameSpan.textContent = sender +": ";
	nameSpan.style.color =  sender === userID? userColor : "#000";

	msgDiv.appendChild(nameSpan);
	msgDiv.appendChild(document.createTextNode(text));
	chatlog.appendChild(msgDiv);
	chatlog.scrollTop = chatlog.scrollHeight;
}

sendBtn.onclick = () => { 
	if (msgInput.value.trim() === "") return;
	ws.send(userID + ": " + msgInput.value);
	msgInput.value = "";
}





