function Replay() {
	this.draw = undefined;
	this.constants = undefined;
	this.robotTypes = undefined;
	this.isReady = true;
	this.requestMore = undefined;
	this.maxRounds = 2000;
	this.Stopped = false;
	this.lastRenderedRound = undefined;
	this.currentRound = 0;
	this.units = {};
	this.effects = {};
	this.effectCount = 0;
	this.lastFrameTime = undefined;

	this.visit = function(response) {
		console.log("visited", response.MessageType);
		if (response.MessageType == "Round") {
			this.parseRound(response.Data);
		} else if (response.MessageType == "StoredConstants") {
			this.parseStoredConstants(response.Data);
		} else if (response.MessageType == "Header") {
			this.parseHeader(response.Data);
		}
		this.isReady = false;
	}

	this.parseRound = function(round) {
		var signals = round.Signals;
		if (signals !== null && signals !== undefined) {
			this.currentRound++;
			for (var i = 0; i < signals.length; i++) {
				var sig = signals[i];
				var type = sig.XMLName.Local;
				if (type === "sig.SpawnSignal") {
					this.processSpawn(sig);
				} else if (type === "sig.MovementSignal") {
					this.processMovement(sig);
				} else if (type === "sig.IndicatorStringSignal") {
				} else if (type === "sig.BroadcastSignal") {
					this.processBroadcast(sig);
				} else if (type === "sig.DeathSignal") {
					this.processDeath(sig);
				} else if (type === "sig.AttackSignal") {
					this.processAttack(sig);
				} else if (type === "sig.InfectionSignal") {
				} else if (type === "sig.TeamResourceSignal") {
				} else if (type === "sig.HealthChangeSignal") {
					this.processHealth(sig);
				} else if (type === "sig.BytecodesUsedSignal") {
				} else if (type === "sig.RobotDelaySignal") {
				} else if (type === "sig.ClearRubbleSignal") {
					this.processClearRubble(sig);
				} else if (type === "sig.RubbleChangeSignal") {
					this.processChangeRubble(sig);
				} else if (type === "sig.PartsChangeSignal") {
					this.processChangeParts(sig);
				} else {
					console.log(sig.XMLName.Local, sig);
					this.Stopped = true;
				}
			}
		}
	}

	this.processSpawn = function(sig) {
		var loc = parseLoc(sig.Loc, this.map.origin);
		var id = sig.RobotId;
		var kind = sig.Type;
		var team = sig.Team;
		var draw = this.draw.createUnitDrawables(kind, team, loc);
		var health = this.robotTypes[kind].maxHealth; 
		this.units[id] = {
			team:team,
			loc:loc,
			from:loc,
			kind:kind, 
			creator:sig.ParentId,
			delay:sig.Delay,
			start:this.currentRound, 
			draw:draw,
			health:health,
			maxHealth:health,
			anim:false
		}
		this.draw.drawUnit(this.units, loc.x, loc.y, team, kind, draw.unit);
	}

	this.processMovement = function(sig) {
		var loc = parseLoc(sig.NewLoc, this.map.origin);
		var id = sig.RobotId;
		var robot = this.units[id];
		robot.from = robot.loc;
		robot.loc = loc;
		robot.delay = sig.Delay;
		robot.start = this.currentRound;
		robot.anim = true;
	}

	this.processDeath = function(sig) {
		var robot = this.units[sig.ObjectId];
		this.draw.removeUnitDrawables(robot);
		this.addDeathEffect(robot.loc, this.currentRound, robot.team);
		delete this.units[sig.ObjectId];
	}

	this.processAttack = function(sig) {
		var targetLoc = parseLoc(sig.TargetLoc, this.map.origin);
		var id = sig.RobotId;
		var robot = this.units[id];
		this.addAttackEffect(robot.loc, targetLoc, this.currentRound, robot.team);
	}

	this.processHealth = function(sig) {
		var ids = parse1DArray(sig.RobotIds);
		var healthStats = parse1DArray(sig.Health);
		for (var j = 0; j < ids.length; j++) {
			var robotId = ids[j];
			var health = healthStats[j];
			var robot = this.units[robotId];
			robot.health = health;
			var healthRatio = robot.health / robot.maxHealth;
			this.draw.drawHealthBar(robot.draw.bar, robot.loc, robot.team, healthRatio);
		}
	}

	this.processClearRubble = function(sig) {
		var robot = this.units[sig.RobotId];
		var loc = parseLoc(sig.Loc, this.map.origin);
		var delay = parseInt(sig.Delay);
		this.addClearEffect(robot.loc, loc, this.currentRound, delay);
		this.draw.drawTile(this.map, this.constants, loc.x, loc.y);
	}

	this.processChangeParts = function(sig) {
		var loc = parseLoc(sig.Loc, this.map.origin);
		var amount = parseInt(sig.Amount);
		this.map.parts[loc.y][loc.x] = amount;
		this.draw.drawTile(this.map, this.constants, loc.x, loc.y);
	}

	this.processChangeRubble = function(sig) {
		var loc = parseLoc(sig.Loc, this.map.origin);
		var amount = parseInt(sig.Amount);
		this.map.rubble[loc.y][loc.x] = amount;
		this.draw.drawTile(this.map, this.constants, loc.x, loc.y);
	}

	this.processBroadcast = function(sig) {
		var loc = parseLoc(sig.Component.Location, this.map.origin);
		var radius = parseInt(sig.Radius);
		this.addBroadcastEffect(loc, this.currentRound, radius);
	}

	this.parseStoredConstants = function(storedConstants) {
		this.constants = {}; 
		this.robotTypes = {}; 

		var gcs = storedConstants.GameConstants;
		for (var i = 0; i < gcs.length; i++) {
			var key = gcs[i].Name; 
			var value = parseValue(gcs[i].Value.Data, gcs[i].Value.XMLName.Local);
			this.constants[key] = value;
			console.log(key, value);
		}

		var rts = storedConstants.RobotTypes;
		for (var i = 0; i < rts.length; i++) {
			var key = rts[i].Name; 
			var p = rts[i].Params;
			var robotType = {}
			for (var j = 0; j < p.length; j++) {
				robotType[p[j].Name] = parseValue(p[j].Value.Data, p[j].Value.XMLName.Local);
			}
			this.robotTypes[key] = robotType;
		}
	}

	this.parseHeader = function(header) {
		this.map = {
			width:header.Map.Width,
			height:header.Map.Height,
			name:header.Map.Name,
			origin:parseLoc(header.Map.Origin),
			rubble:parseArray(header.Map.InitialRubble),
			parts:parseArray(header.Map.InitialParts)
		};
		this.draw.resize(this.draw.renderer.width, this.draw.renderer.height, this.map); 
		this.draw.drawMap(this.map, this.constants);
	}

	this.createPixi = function(width, height) {
		this.draw = new Draw(width, height);
		this.draw.loadResources();
	}

	var n = 0;
	var sumdt = 0.0;
	var sampleSize = 100;
	var lastFrameTime = undefined;
	this.animate = function() {
		var now = new Date();
		if (this.lastRenderedRound !== this.currentRound) {
			this.lastRenderedRound = this.currentRound;
		}
		var time = (now - this.lastRoundTime) / this.draw.MS_PER_ROUND;
		var dt = now - lastFrameTime;
		lastFrameTime = now;
		if (n == sampleSize) {
			var fps = 1000 * sampleSize / sumdt;
			console.log(fps);
			n = 0;
			sumdt = 0.0;
		} else {
			n++;
			sumdt += dt;
		}
		if (time < 0 || time >= 1) {
			time = Math.max(0, Math.min(1, time));
		}
	
		if (this.map !== undefined) {
			for (var j = 0; j < this.map.height; j++) {
				for (var i = 0; i < this.map.width; i++) {
					var tile = this.draw.tiles[j][i];
					//tile.pivot.x = 0.5;
					//tile.pivot.y = 0.5;
					//tile.rotation += 0.1;
				}
			}
		}
		this.draw.drawUnits(this.units, this.currentRound, time);
		this.draw.drawEffects(this.effects, this.currentRound, time);
		
		if (this.draw.bunny !== undefined) {
			this.draw.bunny.rotation += 0.1;
		}

		if (this.draw.mapHasChanged) {
			this.draw.map.cacheAsBitmap = false;
			this.draw.renderer.render(this.draw.stage);
			this.draw.map.cacheAsBitmap = true;
			this.draw.mapHasChanged = false;
		} else {
			this.draw.renderer.render(this.draw.stage);
		}

		//// swap the buffers ...
		//var temp = this.renderTexture;
		//this.renderTexture = this.renderTexture2;
		//this.renderTexture2 = temp;

		//// set the new texture
		//this.draw.outputSprite.texture = this.draw.renderTexture;
		//this.draw.outputSprite.scale.x += 0.1;
		//
		//// render the stage to the texture
		//// the 'true' clears the texture before the content is rendered
		//this.draw.renderTexture2.render(this.draw.stage, null, false);
	
		this.lastFrameTime = now;
		if (this.Stopped !== true) {
			var game = this;
			requestAnimationFrame(function(){game.animate();});
			if (this.lastRoundTime === undefined) {
				this.lastRoundTime = now;
			}
			var roundDelta = now - this.lastRoundTime;
			if (this.isReady === false && roundDelta > this.draw.MS_PER_ROUND && this.currentRound < this.maxRounds) {
				this.lastRoundTime = now;
				this.requestMore();
			}
		}
	}

	this.addDeathEffect = function(loc, currentRound, team) {
		this.effects[this.effectCount] = {type:"death", x:loc.x, y:loc.y, duration:1, start:currentRound, team:team}; 
		this.effectCount++;
	}

	this.addBroadcastEffect= function(loc, currentRound, radius) {
		this.effects[this.effectCount] = {type:"broadcast", x:loc.x, y:loc.y, duration:4, start:currentRound, radius:radius}; 
		this.effectCount++;
	}

	this.addAttackEffect = function(sourceLoc, targetLoc, currentRound, team) {
		this.effects[this.effectCount] = {type:"attack", x:sourceLoc.x, y:sourceLoc.y, x2:targetLoc.x, y2:targetLoc.y, duration:1, start:currentRound, team:team}; 
		this.effectCount++;
	}

	this.addClearEffect = function(sourceLoc, targetLoc, currentRound, delay) {
		this.effects[this.effectCount] = {type:"clear", x:sourceLoc.x, y:sourceLoc.y, x2:targetLoc.x, y2:targetLoc.y, duration:delay, start:currentRound}; 
		this.effectCount++;
	}

	return this;
}

function parse1DArray(raw) {
	var line = raw.split(',');
	var a = new Array(line.length)
	for (var i = 0; i < line.length; i++) {
		a[i] = parseFloat(line[i]);
	}
	return a
	
}

function parseArray(raw) {
	var a = new Array(raw.length);
	for (var i = 0; i < a.length; i++) {
		var line = raw[i][0].split(',');
		a[i] = new Array(line.length)
		for (var j = 0; j < a[i].length; j++) {
			a[i][j] = parseFloat(line[j]);
		}
	}
	return a
	
}

function parseLoc(raw, origin) {
	var s = raw.split(",");
	var ox = 0;
	var oy = 0;
	if (origin !== undefined) {
		ox = origin.x;
		oy = origin.y;
	}
	return {x:parseFloat(s[0])-ox, y:parseFloat(s[1])-oy};
}

function rgb(r, g, b) {
	r = Math.min(255,Math.max(0,Math.floor(r*256)));
	g = Math.min(255,Math.max(0,Math.floor(g*256)));
	b = Math.min(255,Math.max(0,Math.floor(b*256)));
	var color = r*256*256 + g*256 + b;
	return color;
}


function teamColor(team, b) {
	if (b === undefined) {
		b = 0;
	}
	var color;
	if (team === "A") {
		color = rgb(b+1,b,b);
	} else if (team === "B") {
		color = rgb(b,b,b+1);
	} else if (team === "NEUTRAL") {
		color = rgb(b+1,b+1,b+1);
	} else {
		color = rgb(b,b+1,b);
	}
	return color;
}

function Draw(width, height) {
	console.log("Screen size", width, height);

	// You can use either `new PIXI.WebGLRenderer`, `new PIXI.CanvasRenderer`, or `PIXI.autoDetectRenderer`
	// which will try to choose the best renderer for the environment you are in.
	this.renderer = new PIXI.WebGLRenderer(width, height);
	//this.renderer.resolution = 100;
	
	// The renderer will create a canvas element for you that you can then insert into the DOM.
	document.body.appendChild(this.renderer.view);
	
	// You need to create a root container that will hold the scene you want to draw.
	this.stage = new PIXI.Container();
		
	this.sprites = {
		'ZOMBIEDEN':'images/zombieden.png',
		'TTM':'images/ttm.png',
		'TURRET':'images/turret.png',
		'VIPER':'images/viper.png',
		'GUARD':'images/guard.png',
		'SOLDIER':'images/soldier.png',
		'SCOUT':'images/scout.png',
		'ARCHON':'images/archon.png',
		'RANGEDZOMBIE':'images/rangedzombie.png',
		'STANDARDZOMBIE':'images/standardzombie.png',
		'FASTZOMBIE':'images/fastzombie.png',
		'BIGZOMBIE':'images/bigzombie.png',
		'OTHER':'images/other.png'
	};

	// create two render textures... these dynamic textures will be used to draw the scene into itself
	//this.renderTexture = new PIXI.RenderTexture(this.renderer, this.renderer.width, this.renderer.height);
	//this.renderTexture2 = new PIXI.RenderTexture(this.renderer, this.renderer.width, this.renderer.height);

	// create a new sprite that uses the render texture we created above
	//this.outputSprite = new PIXI.Sprite(this.renderTexture);
	
	// align the sprite
	//this.outputSprite.position.x = 400;
	//this.outputSprite.position.y = 300;
	//this.outputSprite.anchor.set(0.5);

	// add to stage
	//this.stage.addChild(this.outputSprite);

	this.board = new PIXI.Container();

	this.map = new PIXI.Container();
	this.board.addChild(this.map);
	
	this.tiles = undefined;

	this.units = new PIXI.Graphics();
	this.units.x=0.5;
	this.units.y=0.5;
	this.board.addChild(this.units);

	this.effects = new PIXI.Graphics();
	this.effects.x=0.5;
	this.effects.y=0.5;
	this.board.addChild(this.effects);

	this.stage.addChild(this.board);
			
	this.MS_PER_ROUND = 200;
	this.mapHasChanged = false;
	this.scale = undefined;
	
	this.resize = function(width, height, map) {
		var margin = 10;
		var sx = (this.renderer.width - margin*2) / map.width;
		var sy = (this.renderer.height - margin*2) / map.height;
		this.scale = Math.min(sx, sy);
		var n = 1.0;
		this.ns = this.scale / n;
		this.nsi = 1.0 / this.ns; 
		var mx = (this.renderer.width - map.width * this.scale) / 2;
		var my = (this.renderer.height - map.height * this.scale) / 2;
	
		this.board.x = mx;
		this.board.y = my;
		this.board.scale.x = this.scale;
		this.board.scale.y = this.scale;
		
		this.map.scale.x = this.nsi;
		this.map.scale.y = this.nsi;
		this.map.x = 0;
		this.map.y = 0;
		this.createMapTiles(map);
	}

	this.loadResources = function() {
		var draw = this;
		// load the texture we need
		PIXI.loader.add('bunny', 'https://raw.githubusercontent.com/pixijs/pixi.js/master/test/textures/bunny.png').load(function (loader, resources) {
			// This creates a texture from a 'bunny.png' image.
			var bunny = new PIXI.Sprite(resources.bunny.texture);
			draw.bunny = bunny;
		
			// Setup the position and scale of the bunny
			bunny.position.x = draw.renderer.width-25;
			bunny.position.y = 300;
		
			bunny.anchor.x = 0.5;
			bunny.anchor.y = 0.5;
		
			// Add the bunny to the scene we are building.
			draw.stage.addChild(bunny);
		});
	
		for (var id in this.sprites) {
			if (this.sprites.hasOwnProperty(id)) {
				PIXI.loader.add(id, this.sprites[id]).load();
			}
		}
	}

	this.drawHealthBar = function(g, loc, team, healthRatio) {
		g.clear();
		g.lineStyle(0, 0x000000, 1);
		g.beginFill(teamColor(team, 0.5), 1);
		g.drawRect(0, 0.9, healthRatio, 0.1);
		g.endFill();
	}

	this.createUnitDrawables = function(kind, team, loc) {
		var resource = this.sprites[kind];
		if (resource === undefined) {
			resource = this.sprites['OTHER'];
		}
		var sprite = PIXI.Sprite.fromImage(resource);
		var color = teamColor(team);
		sprite.tint = color;
		sprite.width = 1;
		sprite.height = 1;

		var bar = new PIXI.Graphics();
		bar.width = 1;
		bar.height = 1;
		this.drawHealthBar(bar, loc, team, 1);
		
		var unit = new PIXI.Container();
		unit.width = 1;
		unit.height = 1;

		unit.addChild(sprite);
		unit.addChild(bar);
		this.units.addChild(unit);

		return {bar:bar, sprite:sprite, unit:unit};
	}

	this.removeUnitDrawables = function(robot) {
		this.units.removeChild(robot.draw.unit);
	}

	this.createMapTiles = function(map) {
		this.tiles = new Array(map.height);
		for (var j = 0; j < map.height; j++) {
			this.tiles[j] = new Array(map.width);
			for (var i = 0; i < map.width; i++) {
				var tile = new PIXI.Graphics();
				tile.x = i * this.ns;
				tile.y = j * this.ns;
				tile.scale.x = this.ns;
				tile.scale.y = this.ns;
				this.map.addChild(tile);
				this.tiles[j][i] = tile; 
			}
		}
	}

	this.drawTile = function(map, constants, x, y) {
		var RT = constants.RUBBLE_OBSTRUCTION_THRESH;
		var PT = 75;
		var g = this.tiles[y][x];
		g.clear();
		
		// draw rubble
		var rubble = map.rubble[y][x];
		var color;
		var alpha;
		if (rubble > RT) {
			color = 0xCCCCCC;
			alpha = 0.5;
		} else {
			color = 0xAAAAAA;
			alpha = rubble * 0.5 / RT;
		}
		g.lineStyle(2/this.scale, 0x555555, 0.1);
		g.beginFill(color, alpha);
		g.drawRect(0,0, 1,1);
		g.endFill();

		// draw parts
		var part = map.parts[y][x];
		if (part > 1) {
			var color = 0xCC7A00;
			g.lineStyle(0, 0x555500, 1);
			g.beginFill(color, 1.0);
			var radius = Math.min(0.4, part / PT);
			g.lineStyle(1, color, 0);
			g.beginFill(color, 1);
			g.drawCircle(0.5,0.5, radius);
			g.endFill();
		}
		this.mapHasChanged = true;
	}

	this.drawMap = function(map, constants) {
		for (var j = 0; j < map.height; j++) {
			for (var i = 0; i < map.width; i++) {
				this.drawTile(map, constants, i, j);
			}
		}
	}
	
	this.drawUnits = function(units, currentRound,  time) {
		for (var id in units) {
			if (units.hasOwnProperty(id)) {
				var robot = units[id];
				if (robot.anim) {
					var t = Math.min(1, currentRound - robot.start + time * 2);
					var x = robot.from.x*(1-t) + robot.loc.x * t;
					var y = robot.from.y*(1-t) + robot.loc.y * t;
					this.drawUnit(this.units, x, y, robot.team, robot.kind, robot.draw.unit);
					if (t == 1) {
						robot.anim = false;
					}
				}
			}
		}
	}

	this.drawUnit = function(graphic, x, y, team, kind, sprite) {
		sprite.x = x-0.5;
		sprite.y = y-0.5;
	}

	this.drawAttack = function(g, x, y, x2, y2, t, team) {
		var t1 = Math.min(1,Math.max(0, (t)));
		var t0 = Math.min(1,Math.max(0, t - 0.1));
		var sx = x*(1-t0) + x2*t0;
		var sy = y*(1-t0) + y2*t0;
		var ex = x*(1-t1) + x2*t1;
		var ey = y*(1-t1) + y2*t1;
		var color = teamColor(team, 0.5); 
		g.lineStyle(2/this.scale, color, 1);
		g.moveTo(sx, sy);
		g.lineTo(ex, ey);
		g.endFill();
	}

	this.drawDeath = function(g, x, y, t, team) {
		var r = Math.min(1,Math.max(0, t));
		var color = teamColor(team, 0.5); 
		g.lineStyle(0, color, 1);
		g.beginFill(color, 1);
		g.drawCircle(x, y, r/2);
		g.endFill();
	}

	this.drawBroadcast = function(g, x, y, t) {
		var r = Math.min(2.1,Math.max(0.1, 0.1 + t*2));
		var alpha = Math.min(1,Math.max(0, 1-t));
		var color = 0xFF00FF; 
		g.lineStyle(1.0/this.scale, color, alpha);
		g.beginFill(color, 0);
		g.drawCircle(x, y, r/2);
		g.endFill();
	}

	this.drawEffects = function(effects, currentRound, time) {
		this.effects.clear();
		var deletedEffects = {};
		for (var id in effects) {
			if (effects.hasOwnProperty(id)) {
				var effect = effects[id]
				if (currentRound - effect.start >= effect.duration) {
					deletedEffects[id] = true;
				} else {
					if (effect.type == "attack") {
						var t = (currentRound - effect.start + time) / effect.duration;
						this.drawAttack(this.effects, effect.x, effect.y, effect.x2, effect.y2, t, effect.team);
					} else if (effect.type == "death") {
						var t = (currentRound - effect.start + time) / effect.duration;
						this.drawDeath(this.effects, effect.x, effect.y, t, effect.team);
					} else if (effect.type == "clear") {
						var t = (currentRound - effect.start + time) / effect.duration;
						this.drawAttack(this.effects, effect.x, effect.y, effect.x2, effect.y2, t, "NEUTRAL");
					} else if (effect.type == "broadcast") {
						var t = (currentRound - effect.start + time) / effect.duration;
						this.drawBroadcast(this.effects, effect.x, effect.y, t);
					} else {
						console.log("Unexpected type", effect.type);
					}
				}
			}
		}
		for (var id in deletedEffects) {
			if (deletedEffects.hasOwnProperty(id)) {
				delete effects[id];
			}
		}
	}


	return this;
}

function parseValue(val, kind) {
	if (kind === "int") {
		return parseInt(val);
	} else if (kind === "double") {
		return parseFloat(val);
	} else if (kind === "string") {
		return val;
	} else if (kind === "boolean") {
		return val === "true";
	} else if (kind === "null") {
		return null;
	} else if (kind === "battlecode.common.RobotType") {
		return val;
	} else {
		console.log("Unknown value kind", kind, val);
		return val;
	}	
}

function startReplay() {
	var replayStream = new WebSocket("ws://localhost:8080/replay/stream?id=" + query.id); 
	var game = new Replay();
	var width = window.innerWidth;
	var height = window.innerHeight;
	game.createPixi(width, height);
	game.animate();
	
	replayStream.onopen = function() {
		console.log("Connection Open");
		replayStream.send("");
		game.requestMore = function() {
			isReady = true;
			replayStream.send("")
		}
	}
	 
	replayStream.onmessage = function(e) {
		var response = JSON.parse(e.data);
		game.visit(response);
	};
	
	replayStream.onclose = function() {
		console.log("websocket closed")
		game.Stopped = true;
	}
	
}

startReplay();
