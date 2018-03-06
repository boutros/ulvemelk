const pjax = {

	// STATE

	_pos: 0,
	_states: {},

	// PUBLIC API

	// setup adds event listeners on all link and form elements with a 'data-pjax' attribute.
	// The event will hijack clicks & form submits, fetching the html responses manually and
	// placing it on the element matching the selector in the 'data-pjax' attribute.
	setup: function () {

		// Register events on the whole document
		this._registerEvents(document)

		window.onpopstate = this._handlePopState.bind(this);

		this._states[this._pos] = {
			url: window.location.href,
			title: document.title,
			container: "body",
			fragment: document.body.innerHTML,
			inputs: this._inputValuePairs()
		}
	},

	// PRIVATE METHODS

	_inputValuePairs() {
		const res = []
		const inputs = document.querySelectorAll("input")
		for (let i = 0; i < inputs.length; i++) {
			if (inputs[i].id) {
				res.push(["#" + inputs[i].id, inputs[i].value])
			} else {
				// TODO
				// input elements without id
			}
		}
		return res
	},

	_registerEvents: function (element) {

		const links = element.querySelectorAll("a, form")

		for (let i = 0; i < links.length; i++) {
			const target = links[i].getAttribute("data-pjax")
			if (!target) {
				continue
			}
			if (!document.querySelector(target)) {
				continue
			}
			if (links[i].localName === "form") {
				links[i].addEventListener("submit", this._handleSubmit.bind(this))
			} else {
				links[i].addEventListener("click", this._handleClick.bind(this))
			}
		}

	},

	_handlePopState (event) {
		let newPos = 0
		if (event.state && event.state.pos) {
			newPos = event.state.pos
		}
		const state = this._states[newPos]
		const el = document.querySelector(state.container)
		el.innerHTML = state.fragment
		for (let i = 0; i < state.inputs.length; i++) {
			document.querySelector(state.inputs[i][0]).value = state.inputs[i][1]
		}
		this._registerEvents(el)
	},

	_handleClick (event) {
		event.preventDefault()

		const url = event.target.href
		this._fetch(url, event.target.getAttribute("data-pjax"))
	},

	_handleSubmit (event) {
		event.preventDefault()

		let url = event.target.action + "?"
		const inputs = [].slice.call(event.target.getElementsByTagName("input"));
		inputs.forEach(input => {
		  url = url + encodeURIComponent(input.name) + "=" + encodeURIComponent(input.value) + "&"
		})
		url = url.slice(0, -1)

		this._fetch(url, event.target.getAttribute("data-pjax"))
	},

	_fetch (url, container) {
		const init = {
			method: "GET",
			headers: {
				"Content-Type": "text/html"
			}
		}
		fetch(url, init)
		.then(response => {
			return response.text()
		}).then(html => {
			const page = document.createElement("html");
			page.innerHTML = html
			const fragment = page.querySelector(container)
			if (!fragment) {
				// TODO
				console.log("no fragment")
				return null
			}
			const el = document.querySelector(container)
			el.parentNode.replaceChild(fragment, el)
			this._pos++
			this._states[this._pos] = {
				url: url,
				title: "bla",
				container: container,
				fragment: fragment.innerHTML,
				inputs: this._inputValuePairs()
			}

			window.history.pushState({pos: this._pos}, "", url)
		})
	}
}

window.pjax = pjax