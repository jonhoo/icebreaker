if (!self.fetch) {
	throw new Error("Browser does not support fetch(); manual refresh required");
}

var to = window.location.pathname.replace("/room/", "/poll/");
var since = document.getElementsByClassName("question").length;
var newdata = 0;
var shown = true;
var title = document.title;

var notifyNewQuestion = function(text) {
	if (Notification.permission === "granted") {
		new Notification("A new question was posted",
			{
				body: text,
				badge: "/static/new-question.ico",
				icon: "/static/new-question.ico",
			}
		)
	}
}

if (("Notification" in window)) {
	if (Notification.permission !== "granted" && Notification.permission !== "denied") {
		Notification.requestPermission(function(permission) {
			if (permission !== "granted") {
				notifyNewQuestion = function() {}
			}
		})
	}
} else {
	notifyNewQuestion = function() {}
}

function poll() {
	fetch(to + "?since=" + since).then(function(res) {
		if (!res.ok) {
			res.text().then(function (txt) {
				console.error("Fetch failed: " + txt);
			});
			return;
		}

		if (res.status == 204) {
			return poll();
		}

		res.json().then(function (res) {
			res.forEach(function(q) {
				var qs = document.querySelectorAll(".question");
				var qe = qs[qs.length-1].cloneNode(true);

				if (!q.by_instructor) {
					qe.classList.remove("instructor");
				}

				if (q.author_nic == "") {
					qe.querySelector("footer").remove();
				} else {
					qe.querySelector("footer > cite").textContent = q.author_nic;
				}

				qe.getElementsByTagName("p")[0].textContent = q.text;
				qs[0].parentNode.insertBefore(qe, qs[0]);
				newdata += 1;

				notifyNewQuestion(q.text)
			});

			if (!shown) {
				document.getElementById("icon").href = "/static/new-question.ico";
			}
			if (newdata != 0) {
				document.title = "(" + newdata + ") " + title;
			}

			since = document.getElementsByClassName("question").length;
			poll();
		});
	});
}

window.addEventListener("focus", function(e) {
	if (newdata != 0) {
		document.getElementById("icon").href = "/static/no-question.ico";
		// show (newdata) in window title to give notification bubble
		document.title = title;
		newdata = 0;
	}
	shown = true
}, false);

window.addEventListener("blur", function(e) {
	shown = false
}, false);

poll();
