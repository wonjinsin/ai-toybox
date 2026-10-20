'use strict';
const $ = id => document.getElementById(id);
const planets = [
  {id:'cinder', name:'Cinder', period:12, phase:15, radius:4.8, size:.38, color:0xe58b60},
  {id:'azure', name:'Azure', period:24, phase:135, radius:8, size:.65, color:0x7dafc3},
  {id:'ochre', name:'Ochre', period:48, phase:255, radius:11.6, size:.88, color:0xd4bb7d}
];
const state = {time:0, speed:1, selectedPlanet:null, viewMode:'overview', playing:true};
const storageKey = 'orrery-observations-v1';
let snapshots = [];
const angleAt = (planet,time) => (planet.phase + time * 360 / planet.period) % 360;
const notice = (text,error=false) => { $('notice').textContent=text; $('notice').classList.toggle('error',error); };
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x0b1012);
const camera = new THREE.PerspectiveCamera(42,1,.1,500);
const renderer = new THREE.WebGLRenderer({antialias:true});
renderer.setPixelRatio(Math.min(devicePixelRatio,2));
renderer.outputEncoding = THREE.sRGBEncoding;
$('scene').appendChild(renderer.domElement);
const controls = new THREE.OrbitControls(camera,renderer.domElement);
controls.enablePan = false;
controls.enableDamping = false;
controls.minDistance=3;
controls.maxDistance=85;
controls.maxPolarAngle=Math.PI*.92;
scene.add(new THREE.AmbientLight(0xc6d7ee,.7));
const sunlight=new THREE.PointLight(0xffe4ba,1.8,0,1);
scene.add(sunlight);
const sun=new THREE.Mesh(new THREE.SphereGeometry(1.35,48,32),new THREE.MeshBasicMaterial({color:0xf4bd72}));
scene.add(sun);
const sunWire=new THREE.Mesh(new THREE.SphereGeometry(1.37,24,14),new THREE.MeshBasicMaterial({color:0xffe3ab,wireframe:true,transparent:true,opacity:.12}));
scene.add(sunWire);
const glowCanvas=document.createElement('canvas');
glowCanvas.width=glowCanvas.height=128;
const glowContext=glowCanvas.getContext('2d');
const glow=glowContext.createRadialGradient(64,64,0,64,64,64);
glow.addColorStop(0,'rgba(255,193,105,.35)');glow.addColorStop(.3,'rgba(255,193,105,.15)');glow.addColorStop(1,'rgba(255,193,105,0)');
glowContext.fillStyle=glow;glowContext.fillRect(0,0,128,128);
const halo=new THREE.Sprite(new THREE.SpriteMaterial({map:new THREE.CanvasTexture(glowCanvas),transparent:true,depthWrite:false,blending:THREE.AdditiveBlending}));
halo.scale.set(7,7,1);scene.add(halo);
const starPositions=[];
let seed=42;
const random=()=>{seed=(seed*1664525+1013904223)>>>0;return seed/4294967296;};
for(let i=0;i<700;i++){const az=random()*Math.PI*2;const y=random()*2-1;const r=80+random()*80;const xz=Math.sqrt(1-y*y)*r;starPositions.push(Math.cos(az)*xz,y*r,Math.sin(az)*xz);}
const starGeometry=new THREE.BufferGeometry();starGeometry.setAttribute('position',new THREE.Float32BufferAttribute(starPositions,3));
scene.add(new THREE.Points(starGeometry,new THREE.PointsMaterial({color:0x80968e,size:.10,transparent:true,opacity:.55})));
planets.forEach(planet=>{
  planet.mesh=new THREE.Mesh(new THREE.SphereGeometry(planet.size,40,28),new THREE.MeshStandardMaterial({color:planet.color,roughness:.9,metalness:.05}));
  planet.mesh.userData.planetId=planet.id;scene.add(planet.mesh);
  const points=[];for(let i=0;i<192;i++){const a=i/192*Math.PI*2;points.push(new THREE.Vector3(Math.cos(a)*planet.radius,0,Math.sin(a)*planet.radius));}
  scene.add(new THREE.LineLoop(new THREE.BufferGeometry().setFromPoints(points),new THREE.LineBasicMaterial({color:0x445448,transparent:true,opacity:.75})));
  const ring=new THREE.Mesh(new THREE.TorusGeometry(planet.size*1.45,.035,8,64),new THREE.MeshBasicMaterial({color:0xd7ef95}));
  ring.rotation.x=Math.PI/2;ring.visible=false;planet.mesh.add(ring);planet.highlight=ring;
});
const targetDelta=new THREE.Vector3();
const nextTarget=new THREE.Vector3();
function updateScene(){
  planets.forEach(p=>{const angle=angleAt(p,state.time)*Math.PI/180;p.mesh.position.set(Math.cos(angle)*p.radius,0,Math.sin(angle)*p.radius);p.highlight.visible=state.selectedPlanet===p.id;});
  if(state.viewMode==='tracking')nextTarget.copy(planets.find(p=>p.id===state.selectedPlanet).mesh.position);else nextTarget.set(0,0,0);
  targetDelta.copy(nextTarget).sub(controls.target);camera.position.add(targetDelta);controls.target.copy(nextTarget);controls.update();
  $('time-readout').textContent=state.time.toFixed(2).padStart(6,'0');
  $('timeline').value=state.time;
  const selected=planets.find(p=>p.id===state.selectedPlanet);
  $('planet-info').replaceChildren();
  const name=document.createElement('b');name.textContent=selected?selected.name+' / SELECTED':'No planet selected';
  const info=document.createElement('p');info.textContent=selected?`${selected.period} simulation seconds / orbit · ${angleAt(selected,state.time).toFixed(2)}°`:'Choose a world above or directly in the scene.';
  $('planet-info').append(name,info);
}
function standardFrame(){
  const distance=state.viewMode==='tracking'?7:33*Math.max(1,1/camera.aspect);
  camera.position.copy(controls.target).add(new THREE.Vector3(.30,.72,.77).normalize().multiplyScalar(distance));
  controls.update();
}
function syncControls(){
  $('play').textContent=state.playing?'Pause':'Resume';
  $('play-status').textContent=state.playing?'LIVE SIMULATION':state.time===120?'END OF TIMELINE':'SIMULATION PAUSED';
  $('mode-status').textContent=state.viewMode==='tracking'?'TRACKING / '+state.selectedPlanet.toUpperCase():'SYSTEM OVERVIEW';
  $('speed-readout').textContent=Number(state.speed.toFixed(6))+'×';$('speed').value=state.speed;
  $('track').disabled=state.selectedPlanet===null;
  $('track').setAttribute('aria-pressed',String(state.viewMode==='tracking'));
  $('overview').setAttribute('aria-pressed',String(state.viewMode==='overview'));
  document.querySelectorAll('[data-planet]').forEach(button=>button.setAttribute('aria-pressed',String(button.dataset.planet===state.selectedPlanet)));
}
let lastTick=performance.now();
function advance(now){
  const elapsed=Math.max(0,(now-lastTick)/1000);lastTick=Math.max(now,lastTick);
  if(state.playing){state.time=Math.min(120,state.time+elapsed*state.speed);if(state.time===120){state.playing=false;syncControls();}}
}
function selectPlanet(id){advance(performance.now());state.selectedPlanet=id;updateScene();syncControls();}
$('play').addEventListener('click',()=>{advance(performance.now());if(state.time>=120){notice('End of timeline. Seek to an earlier time to resume.');return;}state.playing=!state.playing;syncControls();updateScene();});
$('timeline').addEventListener('input',event=>{lastTick=performance.now();state.time=Number(event.target.value);state.playing=false;updateScene();syncControls();});
$('speed').addEventListener('input',event=>{advance(performance.now());state.speed=Number(event.target.value);syncControls();updateScene();});
document.querySelectorAll('[data-planet]').forEach(button=>button.addEventListener('click',()=>selectPlanet(button.dataset.planet)));
$('track').addEventListener('click',()=>{if(!state.selectedPlanet)return;advance(performance.now());state.viewMode='tracking';updateScene();standardFrame();syncControls();});
$('overview').addEventListener('click',()=>{advance(performance.now());state.viewMode='overview';updateScene();standardFrame();syncControls();});
const raycaster=new THREE.Raycaster();const pointer=new THREE.Vector2();
let pointerStart=null;let dragged=false;const activePointers=new Set();
renderer.domElement.addEventListener('pointerdown',event=>{activePointers.add(event.pointerId);if(activePointers.size>1){dragged=true;return;}pointerStart={x:event.clientX,y:event.clientY,id:event.pointerId};dragged=false;});
renderer.domElement.addEventListener('pointermove',event=>{if(pointerStart&&Math.hypot(event.clientX-pointerStart.x,event.clientY-pointerStart.y)>5)dragged=true;});
renderer.domElement.addEventListener('pointercancel',event=>{activePointers.delete(event.pointerId);pointerStart=null;dragged=true;});
renderer.domElement.addEventListener('pointerup',event=>{
  activePointers.delete(event.pointerId);
  if(!pointerStart||dragged||pointerStart.id!==event.pointerId||event.button!==0){if(!activePointers.size)pointerStart=null;return;}
  pointerStart=null;const rect=renderer.domElement.getBoundingClientRect();pointer.set((event.clientX-rect.left)/rect.width*2-1,-(event.clientY-rect.top)/rect.height*2+1);
  raycaster.setFromCamera(pointer,camera);const hit=raycaster.intersectObjects(planets.map(p=>p.mesh),false)[0];if(hit)selectPlanet(hit.object.userData.planetId);
});
function validatePayload(text){
  const root=JSON.parse(text);
  const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
  const owns=(value,key)=>Object.prototype.hasOwnProperty.call(value,key);
  if(!object(root)||!owns(root,'version')||root.version!==1||!owns(root,'snapshots')||!Array.isArray(root.snapshots)||root.snapshots.length>3)throw new Error('Expected version 1 and zero to three snapshots.');
  const ids=new Set();
  return root.snapshots.map((entry,index)=>{
    const fail=()=>{throw new Error(`Snapshot ${index+1} has invalid or missing fields.`);};
    if(!object(entry)||!['id','time','speed','selectedPlanet','viewMode'].every(key=>owns(entry,key)))fail();
    if(typeof entry.id!=='string'||entry.id.length===0||ids.has(entry.id))fail();
    if(typeof entry.time!=='number'||!Number.isFinite(entry.time)||entry.time<0||entry.time>120)fail();
    if(typeof entry.speed!=='number'||!Number.isFinite(entry.speed)||entry.speed<.25||entry.speed>4)fail();
    if(entry.selectedPlanet!==null&&!planets.some(p=>p.id===entry.selectedPlanet))fail();
    if(!['overview','tracking'].includes(entry.viewMode)||(entry.viewMode==='tracking'&&entry.selectedPlanet===null))fail();
    ids.add(entry.id);
    return {id:entry.id,time:entry.time,speed:entry.speed,selectedPlanet:entry.selectedPlanet,viewMode:entry.viewMode};
  });
}
const serialized=()=>JSON.stringify({version:1,snapshots},null,2);
function persist(){
  try{localStorage.setItem(storageKey,serialized());$('storage-notice').textContent='';}
  catch(error){$('storage-notice').textContent='Saving unavailable. Observations work in memory but may not survive reload.';}
}
function restore(entry){
  state.time=entry.time;state.speed=entry.speed;state.selectedPlanet=entry.selectedPlanet;state.viewMode=entry.viewMode;state.playing=false;lastTick=performance.now();
  updateScene();standardFrame();syncControls();notice('Restored '+entry.id+' · playback paused.');
}
function renderSnapshots(){
  const list=$('snapshots');list.replaceChildren();$('capacity').textContent=`${snapshots.length} / 3`;
  if(!snapshots.length){const empty=document.createElement('div');empty.className='empty';empty.textContent='No observations yet. Capture your first moment.';list.append(empty);}
  snapshots.forEach((entry,index)=>{
    const card=document.createElement('article');card.className='snapshot';
    const title=document.createElement('h3');title.textContent=`${String(index+1).padStart(2,'0')} / ${entry.id}`;
    const values=document.createElement('p');values.textContent=`${entry.time} s · ${entry.speed}× · ${entry.selectedPlanet??'no selection'} · ${entry.viewMode}`;
    const angles=document.createElement('p');angles.textContent=planets.map(p=>`${p.name} ${angleAt(p,entry.time).toFixed(2)}°`).join(' / ');
    const actions=document.createElement('div');actions.className='snapshot-actions';
    const restoreButton=document.createElement('button');restoreButton.textContent='Restore';restoreButton.setAttribute('aria-label','Restore observation '+entry.id);restoreButton.onclick=()=>restore(entry);
    const deleteButton=document.createElement('button');deleteButton.textContent='Delete';deleteButton.setAttribute('aria-label','Delete observation '+entry.id);deleteButton.onclick=()=>{snapshots=snapshots.filter(item=>item.id!==entry.id);persist();renderSnapshots();notice('Observation deleted.');$('capture').focus();};
    actions.append(restoreButton,deleteButton);card.append(title,values,angles,actions);list.append(card);
  });
}
let captureSequence=0;
$('capture').addEventListener('click',()=>{
  if(snapshots.length===3){notice('Capacity reached: three observations. Delete one before capturing another.',true);return;}
  advance(performance.now());updateScene();
  let id;do{id='obs-'+Date.now().toString(36)+'-'+(++captureSequence);}while(snapshots.some(entry=>entry.id===id));
  snapshots.push({id,time:state.time,speed:state.speed,selectedPlanet:state.selectedPlanet,viewMode:state.viewMode});persist();renderSnapshots();notice('Moment captured. Live simulation is unchanged.');
});
$('export').addEventListener('click',()=>{$('json').value=serialized();notice('Exported '+snapshots.length+' observations.');});
$('import').addEventListener('click',()=>{
  try{const validated=validatePayload($('json').value);snapshots=validated;persist();renderSnapshots();notice('Imported '+snapshots.length+' observations. Live scene preserved.');}
  catch(error){notice('Import rejected: '+error.message,true);}
});
try{
  const saved=localStorage.getItem(storageKey);
  if(saved!==null){try{snapshots=validatePayload(saved);}catch(error){snapshots=[];notice('Recovery: invalid saved observations were discarded. Live controls are ready.',true);}}
}catch(error){$('storage-notice').textContent='Saving unavailable. Observations work in memory but may not survive reload.';}
const resize=()=>{const host=$('scene');camera.aspect=host.clientWidth/host.clientHeight;camera.updateProjectionMatrix();renderer.setSize(host.clientWidth,host.clientHeight);};
new ResizeObserver(resize).observe($('scene'));
resize();updateScene();standardFrame();syncControls();renderSnapshots();
lastTick=performance.now();
function animate(now){advance(now);updateScene();renderer.render(scene,camera);requestAnimationFrame(animate);}
requestAnimationFrame(animate);
