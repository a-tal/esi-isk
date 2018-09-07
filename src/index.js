import(/* webpackPreload: true */ 'jquery');
import(/* webpackPreload: true */ 'bootstrap');
import(/* webpackPreload: true */ 'bootstrap/dist/css/bootstrap.min.css');
import(/* webpackPreload: true */ 'github-fork-ribbon-css/gh-fork-ribbon.css');

import { pad } from './utils';

function header() {
  let title = document.createElement('h1');
  title.classList.add('display-1');
  title.classList.add('text-center');

  let titleLink = document.createElement('a');
  titleLink.href = 'javascript:window.m()';
  titleLink.classList.add('text-dark');
  titleLink.innerHTML = 'ESI ISK';
  title.append(titleLink);

  let description = document.createElement('h2');
  description.classList.add('text-muted');
  description.classList.add('text-center');
  description.innerHTML = 'Tracking ISK donations and zero ISK item exchange contracts';

  let header = document.createElement('div');
  header.id = "header";
  header.appendChild(title);
  header.appendChild(description);
  header.classList.add('pb-4');
  return header;
}

// create & return the sign in/up div
function signup() {
  let signupLink = document.createElement('a');
  signupLink.id = 'signup';
  signupLink.href = '/signup';
  signupLink.classList.add('github-fork-ribbon');
  signupLink.classList.add('right-top');
  signupLink.innerHTML = 'sign up';
  signupLink.title = 'sign up';
  signupLink.setAttribute('data-ribbon', 'sign up');

  return signupLink;
}

// create & return the top recipients div
function getTop() {
  let top = document.createElement('div');
  top.id = 'top';
  top.classList.add('container');

  let donators = document.createElement('div');
  donators.id = 'donators';
  donators.classList.add('row');

  let recipients = document.createElement('div');
  recipients.id = 'recipients';
  recipients.classList.add('row');

  let recipTitle = document.createElement('h2');
  recipTitle.innerHTML = 'Top Recipients';
  recipTitle.classList.add('col-12');
  recipTitle.classList.add('text-center');
  recipTitle.classList.add('py-4');

  top.appendChild(document.createElement('hr'));
  top.appendChild(recipTitle);
  top.appendChild(recipients);

  let donoTitle = document.createElement('h2');
  donoTitle.innerHTML = 'Top Donators';
  donoTitle.classList.add('col-12');
  donoTitle.classList.add('text-center');
  donoTitle.classList.add('py-4');

  top.appendChild(document.createElement('hr'));
  top.appendChild(donoTitle);
  top.appendChild(donators);

  jQuery.ajax({
    url: "/api/top",
    success: function(t) {
      for (let i = 0; i < t.donators.length; i++) {
        donators.appendChild(characterDiv(
         t.donators[i].id,
         t.donators[i].name,
         t.donators[i].donated_isk || 0
       ))
      }

      for (let i = 0; i < t.recipients.length; i++) {
        recipients.appendChild(characterDiv(
          t.recipients[i].id,
          t.recipients[i].name,
          t.recipients[i].received_isk || 0
        ))
      }
    },
    // XXX handle failure here
  });

  return top;
}

function characterDiv(charID, charName, isk) {
  let div = document.createElement('div');
  div.classList.add('character');
  div.classList.add('col-2');
  div.classList.add('text-center');

  let link = document.createElement('a');
  link.href = "javascript:window.c("+charID+");"

  let char = largeCharacterImage(charID, charName);
  link.appendChild(char);

  let iskLine = document.createElement('p'); // maybe div
  iskLine.classList.add('top-isk');
  iskLine.classList.add('text-muted');
  iskLine.classList.add('text-center');
  iskLine.classList.add('font-weight-light');

  let iskAmount = document.createElement('small');
  iskAmount.innerHTML = formatISK(isk);
  iskLine.appendChild(iskAmount);

  div.appendChild(link);
  div.appendChild(iskLine);

  return div
}

// returns the cleared body div
function clearBodyDiv() {
  let body = document.getElementById('body');
  for (let i = 0; i <= body.children.length; i++) {
    body.children[0].remove();
  }
  return body;
}

// exposed as window.c because reasons
function switchCharacterView(charID) {
  let body = clearBodyDiv();
  body.appendChild(characterViewDiv(charID));
  window.history.pushState(window.history.state, "ESI ISK - " + charID, '/?c=' + charID);
}

// exposed as window.m because reasons
function switchToMainPage() {
  let body = clearBodyDiv();
  body.appendChild(getTop());
  window.history.pushState(window.history.state, "ESI ISK", '/');
}

function frontPage() {
  let body = document.createElement('div');
  body.id = "body";
  body.appendChild(getTop());
  return body;
}

function createTable() {
  let table = document.createElement('table');
  table.classList.add('table');
  table.classList.add('table-hover');
  table.classList.add('table-sm');
  return table;
}

function createTableHeader(content) {
  let th = document.createElement('th');
  th.classList.add('text-center');
  th.innerHTML = content;
  return th;
}


// donations in
function donationsTable() {
  let table = createTable();
  table.id = "donations";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Donator"));
  headerRow.appendChild(createTableHeader("Amount"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Timestamp"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// donations out
function donatedTable() {
  let table = createTable();
  table.id = "donations";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Receiver"));
  headerRow.appendChild(createTableHeader("Amount"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Timestamp"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// contracts in
function contractsTable() {
  let table = createTable()
  table.id = "contracts";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  // XXX ADD CONTRACT ITEMS -- NESTED ON CLICK IDEALLY

  headerRow.appendChild(createTableHeader("Donator"));
  headerRow.appendChild(createTableHeader("Value"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Location"));
  headerRow.appendChild(createTableHeader("Accepted"));
  headerRow.appendChild(createTableHeader("Issued"));
  headerRow.appendChild(createTableHeader("Expires"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function contractedTable() {
  let table = createTable();
  table.id = "contracts";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  // XXX ADD CONTRACT ITEMS -- NESTED ON CLICK IDEALLY

  headerRow.appendChild(createTableHeader("Receiver"));
  headerRow.appendChild(createTableHeader("Value"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Location"));
  headerRow.appendChild(createTableHeader("Accepted"));
  headerRow.appendChild(createTableHeader("Issued"));
  headerRow.appendChild(createTableHeader("Expires"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function dayRow(dt, colSpan) {
  let row = document.createElement('tr');
  let day = createTD(dt.toDateString());
  day.classList.add('table-active');
  day.colSpan = colSpan;
  row.appendChild(day);
  return row;
}

function createTD(content='', centered=true) {
  let td = document.createElement('td');
  td.classList.add('align-middle');
  if (centered) {
    td.classList.add('text-center');
  }
  td.innerHTML = content;
  return td;
}

function donationRow(d, donation) {
  let row = document.createElement('tr');

  let contact = createTD();
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let ts = new Date(d.timestamp);

  row.appendChild(contact);
  row.appendChild(createTD(formatISK(d.amount)));
  row.appendChild(createTD(d.note || '', false));
  row.appendChild(createTD(pad(ts.getUTCHours(), 2) + ':' + pad(ts.getUTCMinutes(), 2) + ':' + pad(ts.getUTCSeconds(), 2)));

  return row;
}

function contractRow(d, donation) {
  let contact = createTD();
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let row = document.createElement('tr');

  let ts = new Date(d.issued);

  row.appendChild(contact);
  row.appendChild(createTD(formatISK(d.value)));
  row.appendChild(createTD(d.note));
  row.appendChild(createTD(d.location));
  row.appendChild(createTD(d.accepted));
  row.appendChild(createTD(pad(ts.getUTCHours(), 2) + ':' + pad(ts.getUTCMinutes(), 2) + ':' + pad(ts.getUTCSeconds(), 2)));
  row.appendChild(createTD(d.expires));

  row.classList.add('contract-collapsed');
  // XXX if this could work that'd be neat
  // row.setAttribute('data-toggle', 'tooltip');
  row.title = 'click to expand/collapse';
  return row
}

function contractItems(d) {
  let itemsTable = createTable();
  itemsTable.classList.add("d-none");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');
  headerRow.classList.add('table-active');

  headerRow.appendChild(createTableHeader("Item"));
  headerRow.appendChild(createTableHeader("Quantity"));

  header.appendChild(headerRow);
  itemsTable.appendChild(header);

  let body = document.createElement('tbody');
  itemsTable.appendChild(body);

  for (let i = 0; i < d.items.length; i++) {
    let itemRow = document.createElement('tr');
    let itemImg = document.createElement('img');

    itemImg.height = 50;
    itemImg.width = 50;
    itemImg.alt = d.items[i].type_id.toString();
    itemImg.classList.add('rounded');
    itemImg.src = 'https://imageserver.eveonline.com/Type/' + d.items[i].type_id + '_64.png';

    let itemTD = createTD();
    itemTD.appendChild(itemImg);
    itemRow.appendChild(itemTD);
    itemRow.appendChild(createTD(d.items[i].quantity));

    body.appendChild(itemRow);
  }

  let itemsTR = document.createElement('tr');
  let itemsTD = document.createElement('td');
  itemsTD.colSpan = 7;
  itemsTD.appendChild(itemsTable);
  itemsTR.appendChild(itemsTD);

  return itemsTR;
}

function formatISK(n) {
  return n.toLocaleString(undefined, {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2
  }) + ' ISK';
}

function largeCharacterImage(charID, charName, imgType, imgExtn) {
  let charImg = document.createElement('img');
  charImg.height = 128;
  charImg.width = 128;
  charImg.alt = charID.toString();
  charImg.classList.add('rounded');
  charImg.src = 'https://imageserver.eveonline.com/' + (imgType || 'Character') + '/' + charID + '_128.' + (imgExtn || 'jpg');

  let charDiv = document.createElement('div');
  let charNameP = document.createElement('p');
  charNameP.classList.add('text-muted');
  charNameP.classList.add('my-0');

  charNameP.innerHTML = charName;
  charDiv.appendChild(charImg);
  charDiv.appendChild(charNameP);

  return charDiv;
}

function smallCharacterImage(charID) {
  let charImg = document.createElement('img');
  charImg.height = 50;
  charImg.width = 50;
  charImg.alt = charID.toString();
  charImg.classList.add('rounded');
  charImg.src = 'https://imageserver.eveonline.com/Character/' + charID + '_64.jpg';
  return charImg;
}

// return title as an h3 with muted and centered text
function h3title(title) {
  let h3 = document.createElement('h3');
  h3.classList.add('text-center');
  h3.classList.add('text-muted');
  h3.innerHTML = title;
  return h3;
}

function characterView(charID) {
  let body = document.createElement('div');
  body.id = "body";
  body.appendChild(characterViewDiv(charID));
  return body
}

function characterViewDiv(charID) {
  let character = document.createElement('div');
  character.classList.add('text-center');
  character.classList.add('pb-4');
  character.classList.add('row');
  character.classList.add('justify-content-center');

  let charImg = largeCharacterImage(charID, '');
  charImg.classList.add('col-2');
  character.appendChild(charImg);

  let details = document.createElement('div');
  details.appendChild(character);

  // in
  let donations = donationsTable();
  let contracts = contractsTable();
  // out
  let donated = donatedTable();
  let contracted = contractedTable();

  jQuery.ajax({
    url: "/api/char?c=" + charID,
    success: function(t) {

      charImg.getElementsByTagName('p')[0].innerHTML = t.character.name;

      let corpImg = largeCharacterImage(
        t.character.corporation,
        t.character.corporation_name,
        'Corporation',
        'png',
      );
      corpImg.classList.add('col-2');
      character.appendChild(corpImg);

      if (t.character.alliance != undefined) {
        let alianceImg = largeCharacterImage(
          t.character.alliance,
          t.character.alliance_name,
          'Alliance',
          'png',
        );
        alianceImg.classList.add('col-2');
        character.appendChild(alianceImg);
      }

      if (t.contracts != undefined) {
        displayTable(details, contracts, "contracts", t.contracts, contractRow, true, 7)
      }

      if (t.donations != undefined) {
        displayTable(details, donations, "donations", t.donations, donationRow, true)
      }

      if (t.contracted != undefined) {
        displayTable(details, contracted, "contracted", t.contracted, contractRow, false, 7)
      }

      if (t.donated != undefined) {
        displayTable(details, donated, "donated", t.donated, donationRow, false)
      }

    },
    // XXX handle failure here
  });

  return details;
}

function displayTable(details, parent, title, array, genFunc, donation, colSpan=4) {
  details.appendChild(h3title(title));
  details.appendChild(parent);
  let day = 0;

  for (let i=array.length-1; i >= 0; i--) {
    let thisDay = new Date(array[i].timestamp || array[i].issued);
    if (thisDay.getUTCDate() != day) {
      day = thisDay.getUTCDate();
      parent.tBodies[0].appendChild(dayRow(thisDay, colSpan));
    }
    parent.tBodies[0].appendChild(genFunc(array[i], donation));
    // XXX HACKHACKHACKHACKHACKHACKHACK XXX
    if (colSpan != 4) {
      parent.tBodies[0].appendChild(contractItems(array[i]));
    }
  }
}

function footer() {
  let copyright = document.createElement('p');
  copyright.innerHTML = 'Source: <a href="https://github.com/a-tal/esi-isk">https://github.com/a-tal/esi-isk</a><br />Released under the MIT license<br />© 2018 Adam Talsma';
  copyright.classList.add('small');
  copyright.classList.add('text-muted');
  copyright.classList.add('text-center');

  let footer = document.createElement('div');
  footer.id = "footer";
  footer.appendChild(document.createElement('hr'));
  footer.appendChild(copyright);
  return footer;
}

function content() {
  let container = document.createElement('div');
  container.id = "container";
  container.className = 'container-fluid';

  container.appendChild(signup());
  container.appendChild(header());

  var urlParams = new URLSearchParams(window.location.search);
  var charID = urlParams.get('c');

  if (charID != undefined) {
    container.appendChild(characterView(charID));
  } else {
    container.appendChild(frontPage());
  }

  container.appendChild(footer());
  return container;
}

jQuery(function($){
  document.body.removeChild(document.getElementById("nojs"));
  document.body.appendChild(content());
  window.c = switchCharacterView;
  window.m = switchToMainPage;

  $(document).on("click", ".contract-collapsed", function(e) {
    let target = e.target;
    while (target.tagName != "TR") {
      target = target.parentElement;
    }
    let table = target.nextSibling.firstElementChild.firstElementChild;
    table.classList.remove('d-none');
    table.classList.add('d-table');
    target.classList.remove("contract-collapsed");
    target.classList.add("contract-expanded");
  });

  $(document).on("click", ".contract-expanded", function(e) {
    let target = e.target;
    while (target.tagName != "TR") {
      target = target.parentElement;
    }
    let table = target.nextSibling.firstElementChild.firstElementChild;
    table.classList.add('d-none');
    table.classList.remove('d-table');
    target.classList.add("contract-collapsed");
    target.classList.remove("contract-expanded");
  })
});
