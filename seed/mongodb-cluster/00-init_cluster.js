rs.initiate({
	_id: "rs0",
	members: [
		{ _id: 0, host: "mongodb-replica-1", priority: 1 },
		{ _id: 1, host: "mongodb-replica-2", priority: 0.5 },
		{ _id: 2, host: "mongodb-replica-3", priority: 0.5 },
	],
});
