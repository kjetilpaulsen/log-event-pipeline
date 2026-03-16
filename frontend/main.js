const statusEl = document.getElementById("status");
const eventsEl = document.getElementById("events");
const maxEvents = 50;

function addEventCard(event) {
    const card = document.createElement("div");
    meta.className = "meta";
    meta.innerHTML = `
    <span class="level">${event.level}</span>
    | ${event.timestamp}
    | ${event.app_name}
    | ${event.function_name}
    `;

    const message = document.createElement("div");
    message.className = "message";
    message.textContent = event.message;

    card.appendChild(meta);
    card.appendChild(message);

    eventsEl.prepend(card);

    while (eventsEl.children.length > maxEvents) {
        eventsEl.removeChild(eventsEl.lastChild);
    }
}

const source = new EventSource("/events");

source.onopen = function () {
    statusEl.textContent = "Connected";
};

source.onmessage = function (evt) {
    const data = JSON.parse(evt.data);
    addEventCard(data);
};

source.onerror = function () {
    statusEl.textContent = "Disconnected. Retrying";
};

