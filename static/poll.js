if (!self.fetch) {
	throw new Error("Browser does not support fetch(); manual refresh required");
}

var timer;
var interval = 5*1000;
var to = window.location.pathname.replace("/room/", "/poll/");
var since = document.getElementsByClassName("question").length;
console.log(to);

function poll() {
	fetch(to + "?since=" + since).then(function(res) {
		if (!res.ok) {
			res.text().then(function (txt) {
				console.error("Fetch failed: " + txt);
			});
			return;
		}

		if (res.status == 204) {
			timer = window.setTimeout(poll, interval);
			return
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
			});
			since = document.getElementsByClassName("question").length;
			timer = window.setTimeout(poll, interval);
		});
	});
}

timer = window.setTimeout(poll, interval);
