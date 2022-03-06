const socket = io();

const PREFERENCE_COLOR = [0, 200, 0];
const MATCHING_COLOR = [0, 20, 190];
const HISTORY_COLOR = "#2ce8bf";
const defaultGroupSize = 5;
// history
let historySteps;
let historyStepIndex = 0;
let historyLines = new Map();
let processHistory = false;
let showHistory = false;

let showPreferences = true;
let showMatching = true;
let groupSize = defaultGroupSize;
// execution duration of the last algorithm run
let ballRadius = 40;
let runDurationText;

let canvasWidth = visualViewport.width, canvasHeight = visualViewport.height;

let maxBallsPerRow;
let ballMargin;

// hash maps, with the keys being the ID of the person
let men = new Map();
let women = new Map();

function renderGui() {
  let guiContainer = createDiv();

  guiContainer.addClass("flex-container");
  guiContainer.position(10, 10);
  guiContainer.style("width", "25%");

  let genDiv = createDiv();
  let inp = createInput(defaultGroupSize, "number");
  inp.size(50);
  inp.input((v) => {
    groupSize = parseInt(inp.value())
    if (isNaN(groupSize)) {
      inp.value(defaultGroupSize);
    }
  });
  let genButton = createButton("generate");
  genButton.mouseClicked(v => {
    showPreferences = true;
    showHistory = false;
    socket.emit("generate", groupSize);
  });

  let matchingDiv = createDiv();
  let startMatchingButton = createButton("start matching");
  startMatchingButton.mouseClicked(v => {
    showPreferences = false;
    processHistory = true;
    socket.emit("stable-match");
  });

  runDurationText = createSpan();
  runDurationText.style("background-color", "white");
  runDurationText.style("border-radius", "5px");

  let replayHistoryButton = createButton("replay history");
  replayHistoryButton.mouseClicked(v => {
    historyLines.clear();
    processHistory = true;
  });

  matchingDiv.child(startMatchingButton);
  matchingDiv.child(runDurationText);
  matchingDiv.child(replayHistoryButton);

  genDiv.child(inp);
  genDiv.child(genButton);

  guiContainer.child(genDiv);
  guiContainer.child(matchingDiv);
}

function setup() {
  createCanvas(canvasWidth, canvasHeight);

  renderGui();

  socket.on("stable-match-history", function (data) {
    historySteps = data.Steps;
    historyLines.clear();
    showHistory = (historySteps != null && men != null && women != null);
  });

  socket.on("stable-match-duration", function (data) {
    runDurationText.elt.innerText = `execution time: ${data}`;
  });

  socket.on("update-matching", function (data) {
    // assign ID of the new wife
    men.get(data.HusbandID).wife = data.WifeID;
    // delete the wife for the newly unmarried man
    men.get(data.UnmarriedID).wife = null;
  });

  socket.on("update-people", function (data) {
    ballMargin = Math.floor(canvasWidth / data.Men.length) - ballRadius;
    maxBallsPerRow = Math.floor(canvasWidth / (ballMargin + ballRadius));
    // iterate men
    men.clear();
    data.Men.forEach(m => {
      let rowPos = (m.ID % maxBallsPerRow);
      let columnPos = (Math.floor(m.ID / maxBallsPerRow));
      m.pos = { x: (rowPos * ballMargin) + ((rowPos + 1) * ballRadius), y: ballRadius + (columnPos * (ballRadius)) };
      men.set(m.ID, m);
    });
    // iterate women
    women.clear();
    data.Women.forEach(w => {
      let rowPos = (w.ID % maxBallsPerRow);
      let columnPos = (Math.floor(w.ID / maxBallsPerRow));
      w.pos = { x: (rowPos * ballMargin) + ((rowPos + 1) * ballRadius), y: (canvasHeight - ballRadius) - (columnPos * (ballRadius)) };
      women.set(w.ID, w);
    });
  });
}

function processHistoryStep(historyStep) {
  let husband = men.get(historyStep.HusbandID)
  let wife = women.get(historyStep.WifeID)

  // add line between newly married couple
  historyLines.set(historyStep.HusbandID, { m: husband.pos, w: wife.pos });
  // delete line of the man who got separated
  historyLines.delete(historyStep.UnmarriedID)
}

function renderHistoryLines() {
  strokeWeight(5);
  stroke(HISTORY_COLOR);
  historyLines.forEach(l => {
    line(l.m.x, l.m.y, l.w.x, l.w.y);
  });
  noStroke();
}

function draw() {
  background(0);

  // render men
  fill(255);
  men.forEach((m, key) => {
    if (showPreferences) {
      // render connections based on male preferences
      stroke(PREFERENCE_COLOR);
      m.Preferences.forEach((p, index) => {
        let w = women.get(index);
        if (w) {
          strokeWeight(1 + p);
          line(m.pos.x, m.pos.y, w.pos.x, w.pos.y);
        }
      });
      noStroke();
    }
    if (showMatching) {
      if (m.wife) {
        let w = women.get(wife);
        stroke(MATCHING_COLOR);
        strokeWeight(5);
        line(m.pos.x, m.pos.y, w.pos.x, w.pos.y);
        noStroke();
      }
    }
    circle(m.pos.x, m.pos.y, ballRadius);
  });

  // render women
  fill(128, 0, 0)
  women.forEach((w, key) => {
    circle(w.pos.x, w.pos.y, ballRadius);
  });

  if (processHistory && historySteps != null) {
    processHistoryStep(historySteps[historyStepIndex]);
    historyStepIndex++;
    if (historyStepIndex >= historySteps.length) {
      processHistory = false
      historyStepIndex = 0;
    }
  }

  if (showHistory) {
    // set low frame rate so the history will play out slowly
    frameRate(2);
    // render history
    renderHistoryLines();
  }
  else {
    frameRate(30);
  }
}
