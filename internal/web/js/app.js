App = Ember.Application.create({
  LOG_TRANSITIONS: true
});

App.Router.map(function() {
  this.route('register', { path: '/' });
  this.route("clone", { path: "/clone/:name" });
  this.route('players', { path: '/players' });
  // put your routes here
});

App.RegisterRoute = Ember.Route.extend({
  model: function() {
    return {};
  }
});

App.PlayersRoute = Ember.Route.extend({
  model: function() {
    return Ember.$.getJSON('http://akusete.local:8080/players').then(function(data){
      return data;
    });
  }
});

App.CloneRoute = Ember.Route.extend({
  model: function(params) {
    return {
      url:params.name
    }
  }
});

App.RegisterController = Ember.Controller.extend({
  actions: {
    register: function() {
      var T = this;
      if (!T.get('frozen')) {
        T.set('frozen', true); 
        Ember.$.post('http://akusete.local:8080/register', {
            name:this.get('name'),
            public_key:this.get('public_key')	   
          }, null, 'json').done(function(response){
  	  console.log(response, T)
          //T.set('clone', response.repo_url);
          T.transitionToRoute('/clone/' + T.get('name')) 
        }).fail(function(response){
  	  if (response.responseJSON !== null) {
  		//TODO:render error
  	  } else {
  		//TODO: render error occurred
  	  }
        }).always(function(){
          T.set('frozen', false); 
        });
      }
    }
  }
});

