module.exports = {
    eventhandler: function (req, res) {
        res.end("Env: " + JSON.stringify(process.env) + "\n");
    
    }
  }
  
