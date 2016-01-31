function Replay() {
	this.draw = undefined;
	this.constants = undefined;
	this.robotTypes = undefined;
	this.isReady = true;
	this.requestMore = undefined;
	this.maxRounds = 5;
	this.Stopped = false;
	this.lastRenderedRound = undefined;
	this.currentRound = 0;
	this.units = {};
	this.effects = {};
	this.effectCount = 0;
	this.lastFrameTime = undefined;

	this.visit = function(response) {
		console.log("visited", game.isReady, response.MessageType);
		if (response.MessageType == "Round") {
			this.parseRound(response.Data);
		} else if (response.MessageType == "StoredConstants") {
			this.parseStoredConstants(response.Data);
		} else if (response.MessageType == "Header") {
			this.parseHeader(response.Data);
		}
		game.isReady = false;
	}

	this.parseRound = function(round) {
		var signals = round.Signals;
		if (signals !== null && signals !== undefined) {
			game.currentRound++;
			for (var i = 0; i < signals.length; i++) {
				var sig = signals[i];
				var type = sig.XMLName.Local;
				if (type === "sig.SpawnSignal") {
					processSpawn(sig);
				} else if (type === "sig.MovementSignal") {
					processMovement(sig);
				} else if (type === "sig.IndicatorStringSignal") {
				} else if (type === "sig.BroadcastSignal") {
				} else if (type === "sig.DeathSignal") {
					processDeath(sig);
				} else if (type === "sig.AttackSignal") {
					processAttack(sig);
				} else if (type === "sig.InfectionSignal") {
				} else if (type === "sig.TeamResourceSignal") {
				} else if (type === "sig.HealthChangeSignal") {
					processHealth(sig);
				} else if (type === "sig.BytecodesUsedSignal") {
				} else if (type === "sig.RobotDelaySignal") {
				} else if (type === "sig.ClearRubbleSignal") {
				} else if (type === "sig.RubbleChangeSignal") {
				} else if (type === "sig.PartsChangeSignal") {
				} else {
					console.log(sig.XMLName.Local, sig);
					game.Stopped = true;
				}
			}
		}
	}

	this.processSpawn = function(sig) {
		var loc = parseLoc(sig.Loc, this.map.origin);
		var id = sig.RobotId;
		var kind = sig.Type;
		var team = sig.Team;
		var sprite = createUnitSprite(kind, team);
		var healthbar = createHealthBar();
		var health = this.robotTypes[kind].maxHealth; 
		game.draw.units.addChild(healthbar);
		game.draw.units.addChild(sprite);
		this.units[id] = {
			team:team,
			loc:loc,
			from:loc,
			kind:kind, 
			creator:sig.ParentId,
			delay:sig.Delay,
			start:this.currentRound, 
			sprite:sprite,
			healthbar:healthbar,
			health:health,
			maxHealth:health
		}
	}

	this.processMovement = function(sig) {
		var loc = parseLoc(sig.NewLoc, this.map.origin);
		var id = sig.RobotId;
		var robot = this.units[id];
		robot.from = robot.loc;
		robot.loc = loc;
		robot.delay = sig.Delay;
		robot.start = this.currentRound;
	}

	this.processDeath = function(sig) {
		var robot = this.units[sig.ObjectId];
		this.draw.units.removeChild(robot.sprite);
		this.draw.units.removeChild(robot.healthbar);
		this.effects[this.effectCount] = {type:"death", x:robot.loc.x, y:robot.loc.y, duration:1, start:this.currentRound, team:robot.team}; 
		this.effectCount++;
		delete this.units[sig.ObjectId];
	}

	this.processAttack = function(sig) {
		var loc = parseLoc(sig.TargetLoc, this.map.origin);
		var id = sig.RobotId;
		var robot = this.units[id];
		var source = robot.loc;
		this.effects[this.effectCount] = {type:"attack", x:source.x, y:source.y, x2:loc.x, y2:loc.y, duration:1, start:this.currentRound, team:robot.team}; 
		this.effectCount++;
	}

	this.processHealth = function(sig) {
		var ids = parse1DArray(sig.RobotIds);
		var healthStats = parse1DArray(sig.Health);
		for (var j = 0; j < ids.length; j++) {
			var robotId = ids[j];
			var health = healthStats[j];
			if (this.units[robotId] !== undefined) {
				this.units[robotId].health = health;
			}
		}
	}

	this.parseStoredConstants = function(storedConstants) {
		this.constants = {}; 
		this.robotTypes = {}; 

		var gcs = storedConstants.GameConstants;
		for (var i = 0; i < gcs.length; i++) {
			var key = gcs[i].String; 
			var value = gcs[i].Value;
			this.constants[key] = value;
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
		var map = header.Map;
		this.map = {
			width:map.Width,
			height:map.Height,
			name:map.Name,
			origin:parseLoc(map.Origin),
			rubble:parseArray(map.InitialRubble),
			parts:parseArray(map.InitialParts)
		};
	}

	this.createPixi = function() {
		var width = window.innerWidth;
		var height = window.innerHeight;
		console.log("Screen size", width, height);
	
		// You can use either `new PIXI.WebGLRenderer`, `new PIXI.CanvasRenderer`, or `PIXI.autoDetectRenderer`
		// which will try to choose the best renderer for the environment you are in.
		var renderer = new PIXI.WebGLRenderer(width, height);
	
		// The renderer will create a canvas element for you that you can then insert into the DOM.
		document.body.appendChild(renderer.view);
	
		// You need to create a root container that will hold the scene you want to draw.
		var stage = new PIXI.Container();
		
		var sprites = {
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
		var board = new PIXI.Container();

		var map = new PIXI.Graphics();
		board.addChild(map);

		var units = new PIXI.Graphics();
		units.x=0.5;
		units.y=0.5;
		board.addChild(units);

		var effects= new PIXI.Graphics();
		effects.x=0.5;
		effects.y=0.5;
		board.addChild(effects);

		stage.addChild(board);
		
		this.draw = {
			renderer:renderer,
			stage:stage,
			map:map,
			units:units,
			effects:effects,
			sprites:sprites,
			MS_PER_ROUND:200,
			board:board,
			isMapDrawn:false
		};
	
		loadResources();
	}

	this.loadResources = function() {
		var game = this;
		// load the texture we need
		PIXI.loader.add('bunny', 'https://raw.githubusercontent.com/pixijs/pixi.js/master/test/textures/bunny.png').load(function (loader, resources) {
			// This creates a texture from a 'bunny.png' image.
			var bunny = new PIXI.Sprite(resources.bunny.texture);
			game.draw.bunny = bunny;
		
			// Setup the position and scale of the bunny
			bunny.position.x = game.draw.renderer.width-25;
			bunny.position.y = 300;
		
			bunny.anchor.x = 0.5;
			bunny.anchor.y = 0.5;
		
			// Add the bunny to the scene we are building.
			game.draw.stage.addChild(bunny);
		});
	
		for (var id in game.draw.sprites) {
			if (game.draw.sprites.hasOwnProperty(id)) {
				PIXI.loader.add(id, game.draw.sprites[id]).load();
			}
		}
	}

	this.drawMap = function() {
		var game = this;
		var margin = 10;
		var width = game.map.width;
		var height = game.map.height; 
		
		var renderer = game.draw.renderer;
		var sx = (renderer.width - margin*2) / width;
		var sy = (renderer.height - margin*2) / height;
		var scale = Math.min(sx, sy);
		game.draw.scale = scale;
		var mx = (renderer.width - width * scale) / 2;
		var my = (renderer.height - height * scale) / 2;
	
		var board = game.draw.board;
		board.x = mx;
		board.y = my;
		board.scale.x = scale;
		board.scale.y = scale;
	
		var graphics = game.draw.map;
		// set a fill and line style
		graphics.clear();
		//graphics.beginFill(0xFF3300);
		graphics.lineStyle(1/scale, 0x555500, 1);
	
		for (var i = 0; i < width; i++) {
			for (var j = 0; j < height; j++) {
				var x = i;
				var y = j;
				var rubble = game.map.rubble[j][i];
				var part = game.map.parts[j][i];
				var color = rgb(rubble / 100, part / 100, 0); 
				graphics.beginFill(color, 0.5);
				graphics.drawRect(x,y, 1,1);
				graphics.endFill();
			}
		}
		graphics.endFill();
	}
	
	this.createHealthBar = function() {
		var g = new PIXI.Graphics();
		return g;
	}

	this.createUnitSprite = function(kind, team) {
		var sprites = game.draw.sprites;
		var resource = sprites[kind];
		if (resource === undefined) {
			resource = sprites['OTHER'];
		}
		var sprite = PIXI.Sprite.fromImage(resource);
		var color = teamColor(team);
		sprite.tint = color;
		return sprite;
	}

	this.drawUnit = function(graphic, x, y, team, kind, sprite, bar, health) {
		sprite.x = x-0.5;
		sprite.y = y-0.5;
		sprite.width = 1;
		sprite.height = 1;

		bar.clear();
		bar.lineStyle(0, 0x000000, 1);
		bar.beginFill(teamColor(team, 0.5), 1);
		bar.drawRect(x-0.5, y+0.45, health, 0.1);
		bar.endFill();
	}

	this.drawAttack = function(graphic, x, y, x2, y2, t, team) {
		var t1 = Math.min(1,Math.max(0, (t)));
		var t0 = Math.min(1,Math.max(0, t - 0.1));
		var sx = x*(1-t0) + x2*t0;
		var sy = y*(1-t0) + y2*t0;
		var ex = x*(1-t1) + x2*t1;
		var ey = y*(1-t1) + y2*t1;
		var color = teamColor(team, 0.5); 
		graphic.lineStyle(2/game.draw.scale, color, 1);
		graphic.moveTo(sx, sy);
		graphic.lineTo(ex, ey);
		graphic.endFill();
	}

	this.drawDeath = function(graphic, x, y, t, team) {
		var r = Math.min(1,Math.max(0, t));
		var color = teamColor(team, 0.5); 
		graphic.lineStyle(0, color, 1);
		graphic.beginFill(color, 1);
		graphic.drawCircle(x, y, r/2);
		graphic.endFill();
	}

	this.drawUnits = function(time) {
		var game = this;
		for (var id in game.units) {
			if (game.units.hasOwnProperty(id)) {
				var robot = game.units[id]
				var t = Math.min(1, game.currentRound - robot.start + time * 2);
				var x = robot.from.x*(1-t) + robot.loc.x * t;
				var y = robot.from.y*(1-t) + robot.loc.y * t;
				var currentHealth = robot.health / robot.maxHealth;
				drawUnit(game.draw.units, x, y, robot.team, robot.kind, robot.sprite, robot.healthbar, currentHealth);
			}
		}
	}

	this.drawEffects = function(time) {
		game.draw.effects.clear();
		var deletedEffects = {};
		for (var id in game.effects) {
			if (game.effects.hasOwnProperty(id)) {
				var effect = game.effects[id]
				if (game.currentRound - effect.start >= effect.duration) {
					deletedEffects[id] = true;
				} else {
					if (effect.type == "attack") {
						var t = (game.currentRound - effect.start + time) / effect.duration;
						drawAttack(game.draw.effects, effect.x, effect.y, effect.x2, effect.y2, t, effect.team);
					} else if (effect.type == "death") {
						var t = (game.currentRound - effect.start + time) / effect.duration;
						drawDeath(game.draw.effects, effect.x, effect.y, t, effect.team);
					} else {
						console.log("Unexpected type", effect.type);
					}
				}
			}
		}
		for (var id in deletedEffects) {
			if (deletedEffects.hasOwnProperty(id)) {
				delete game.effects[id];
			}
		}
	}

	this.animate = function() {
		var game = this;
		var now = new Date();
		if (game.lastRenderedRound !== game.currentRound) {
			game.lastRenderedRound = game.currentRound;
		}
		var time = (now - game.lastRoundTime) / game.draw.MS_PER_ROUND;
		if (time < 0 || time >= 1) {
			time = Math.max(0, Math.min(1, time));
		}
	
		if (game.map !== undefined && game.draw.isMapDrawn === false) {
			game.draw.isMapDrawn = true;
			drawMap();
		}
	
		if (game.units !== undefined) {
			drawUnits(time);
		}
		if (game.effects !== undefined) {
			drawEffects(time);
		}
	
		if (game.draw.bunny !== undefined) {
			game.draw.bunny.rotation += 0.1;
		}
	
		game.draw.renderer.render(game.draw.stage);
	
		game.lastFrameTime = now;
		if (game.Stopped !== true) {
			requestAnimationFrame(animate);
			if (game.lastRoundTime === undefined) {
				game.lastRoundTime = now;
			}
			var roundDelta = now - game.lastRoundTime;
			if (game.isReady === false && roundDelta > game.draw.MS_PER_ROUND) {
				game.lastRoundTime = now;
				game.requestMore();
			}
		}
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
	r = Math.min(255,Math.max(0,r*256));
	g = Math.min(255,Math.max(0,g*256));
	b = Math.min(255,Math.max(0,b*256));
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

//Stream replay
var replayStream = new WebSocket("ws://localhost:8080/replay/stream?id=" + query.id); 
var game = Replay();
game.createPixi();
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

