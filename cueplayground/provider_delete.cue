OpenstackServerDelete: providers: [{
	action: "delete"
	match: {
		type: "object"
		properties: providerType: type: "string"
	}
	parameters: resource_id: {
		required: true
		explicit: type: "string"
	}
	debug: false
}]