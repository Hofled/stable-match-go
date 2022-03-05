const socket = io();

let ballSize = 20;
let canvasWidth = 600, canvasHeight = 600;

let ballsPerRow;
let ballMargin;

// hash maps, with the keys being the ID of the person
let men = new Map();
let women = new Map();

function setup() {
  createCanvas(600, 600);

  socket.on("update-people", function (data) {
    ballsPerRow = Math.floor(canvasWidth / data.Men.length);
    ballMargin = ballsPerRow - ballSize;
    // iterate men
    men.clear();
    data.Men.forEach(m => {
      let rowPos = (m.ID % ballsPerRow);
      m.pos = { x: ballSize + (rowPos * ballMargin) + (rowPos * ballSize), y: ballSize + (Math.floor(m.ID / ballsPerRow)) };
      men.set(m.ID, m);
    });
    // iterate women
    women.clear();
    data.Women.forEach(w => {
      let rowPos = (w.ID % ballsPerRow);
      w.pos = { x: ballSize + (rowPos * ballMargin) + (rowPos * ballSize), y: (canvasHeight - ballSize) - (Math.floor(w.ID / ballsPerRow)) };
      women.set(w.ID, w);
    });
  });
}

function draw() {
  background(0);

  // render men
  fill(255);
  men.forEach((m, key) => {    
    // render connections based on male preferences
    stroke(0, 200, 0);
    m.Preferences.forEach((p, index) => {
      let w = women.get(index);
      if (w) {
        strokeWeight(1 + p);
        line(m.pos.x, m.pos.y, w.pos.x, w.pos.y);
      }
    });
    noStroke();
    circle(m.pos.x, m.pos.y, ballSize);
  });

  // render women
  fill(128, 0, 0)
  women.forEach((w, key) => {
    circle(w.pos.x, w.pos.y, ballSize);
  });
}
