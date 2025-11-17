// generate random username_color
function getRandomColor() {
	const colors = ["#e22", "#2df", "#45f", "#0f0"];  // Убрал "#0f0f", добавил разнообразие
	return colors[Math.floor(Math.random() * colors.length)];
 }
 
 // nickname & color
 const userID = "User-" + Math.floor(Math.random() * 1000);
 const userColor = getRandomColor();
 
 // websocket connect
 const ws = new WebSocket("ws://localhost:8080/ws");
 
 // dom elements
 const chatlog = document.getElementById("chatlog"); // chatlog 
 const msgInput = document.getElementById("msgInput"); // message user input
 const sendBtn = document.getElementById("sendBtn"); 	// user send button [ send]
 const nameInput = document.getElementById("nameInput") // user  name input 
 const chat = document.getElementById("chat"); //chat 

 let username  = null
 
 // receive message (ИСПРАВЛЕНО: добавил event)
 ws.onmessage = (event) => {  //
	try {
	  const msgDiv = document.createElement("div");
	  msgDiv.classList.add("message");
 
	  const data = event.data;
	  const separatorIndex = data.indexOf(": ");  //
	  if (separatorIndex === -1) {
		 console.error("Invalid message format:", data);
		 return;
	  }
 
	  const sender = data.substring(0, separatorIndex);
	  const text = data.substring(separatorIndex + 2);
 
	  const nameSpan = document.createElement("span");
	  nameSpan.classList.add("username");
	  nameSpan.textContent = sender + ": ";
	  const senderColor =  
	  
	  // Цвет: для своего — случайный, для других — чёрный (или генерируй по sender)
	  nameSpan.style.color = sender === userID ? userColor : "#000";
	  
	  msgDiv.appendChild(nameSpan);
	  msgDiv.appendChild(document.createTextNode(text));
	  chatlog.appendChild(msgDiv);
	  chatlog.scrollTop = chatlog.scrollHeight;
	} catch (error) {
	  console.error("Error processing message:", error);
	}
 };
 
 // send message
 sendBtn.onclick = () => {
	const message = msgInput.value.trim();
	if (message === "") return;
	ws.send(userID + ": " + message);
	msgInput.value = "";
 };
 
 // Дополнительно: Enter для отправки
 msgInput.addEventListener("keypress", (e) => {
	if (e.key === "Enter") sendBtn.click();
 });