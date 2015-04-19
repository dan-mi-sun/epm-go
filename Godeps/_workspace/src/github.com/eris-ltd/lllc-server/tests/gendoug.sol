contract gendoug {
	struct permdata {
		uint permnum;
		uint globalval;
	}
	struct userdata {
		mapping (uint => uint) permvals;
		mapping (uint => bool) exclude;
	}

	mapping (address => userdata) users;
	mapping (bytes32 => permdata) perms;
	uint pcount;
	mapping (bytes32 => address) names;

	function gendoug() {
		pcount =  1;
		permdata p = perms["DOUG"];
		userdata u = users[tx.origin];
		p.permnum = pcount;
		p.globalval = 0;

		//Give creator DOUG permissions
		u.permvals[pcount] = 1;
	}

	//NameReg Functionality
	function checkName(bytes32 name) returns (address ret){
		return names[name];
	}

	function register(bytes32 name, address addr) returns (uint ret) {
		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		names[name] = addr;
		return 1;
	}

	//Variable storage Functionality TODO


	//Permission Functionality
	function getPerm(address addr, bytes32 permName) returns (uint ret) {
		permdata p = perms[permName];
		userdata u = users[addr];

		//Check if permName even exists
		if (p.permnum == 0) return 0;

		//Process perm
		if (u.permvals[p.permnum] == 0){
			if (u.exclude[p.permnum]) return u.permvals[p.permnum];
			else return p.globalval;
		}
		else { 
			return u.permvals[p.permnum]; 
		}
	}

	function setPerm(address addr, bytes32 permName, uint value) returns (uint ret) {
		permdata p = perms [permName];
		userdata u = users[addr];

		//Check if permName even exists
		if (p.permnum == 0) return 0;

		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		u.permvals[p.permnum] = value;

		return 1;
	}

	function addPerm(bytes32 permName) returns (uint ret) {
		permdata p = perms[permName];

		//Check that it doesn't already exist
		if (p.permnum != 0) return 0;

		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		pcount += 1;
		p.permnum = pcount;
		return 1;
	}

	function rmPerm(bytes32 permName) returns (uint ret) {
		permdata p = perms[permName];

		//Check that it doesn't already not exist
		if (p.permnum == 0) return 0;

		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		p.permnum = 0;
		return 1;
	}

	function setGlobal(bytes32 permName, uint value) returns (uint ret) {
		permdata p = perms[permName];

		//Check that it even exists
		if (p.permnum == 0) return 0;

		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		p.globalval = value;

		return 1;
	}

	function setExclude(address addr, bytes32 permName, bool value) returns (uint ret) {
		permdata p = perms[permName];
		userdata u = users[addr];

		//Check that it even exists
		if (p.permnum == 0) return 0;

		//Check for doug perms for ORIGIN
		if (getPerm(tx.origin, "DOUG")!=1) return 0;

		u.exclude[p.permnum] = value;
		return 1;
	}
}
