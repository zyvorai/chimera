// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package gateway

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<meta name="color-scheme" content="light" />
<title>Chimera — Infrastructure Simulation Engine</title>
<style>
:root{
  --bg:#ffffff;--bg-secondary:#f5f5f7;--bg-hero:#fafafa;
  --line:#d2d2d7;--line-light:#e8e8ed;--text:#1d1d1f;--text-secondary:#6e6e73;--text-tertiary:#86868b;
  --link:#0066cc;--accent:#0071e3;--accent-hover:#0077ed;--success:#34c759;--warning:#ff9500;--danger:#ff3b30;
  --shadow:0 2px 16px rgba(0,0,0,.06);--shadow-lg:0 8px 40px rgba(0,0,0,.1);
  --radius-sm:8px;--radius:12px;--radius-lg:18px;--radius-pill:980px;
  --nav-blur:rgba(251,251,253,.8);--content-max:980px;--zyvor-orange:#ff5a15;
}
*{box-sizing:border-box}html{height:100%;scroll-behavior:smooth}body{margin:0;min-height:100%;background:var(--bg);color:var(--text);font:17px/1.47059 -apple-system,BlinkMacSystemFont,"SF Pro Text","SF Pro Display",system-ui,sans-serif;-webkit-font-smoothing:antialiased}
button,input,select{font:inherit;color:inherit}button{cursor:pointer}a{color:var(--link);text-decoration:none}a:hover{text-decoration:underline}.empty-msg{color:var(--text-tertiary)}
.shell{display:none;min-height:100vh;flex-direction:column}#appRoot.show{display:flex}
.login-page{display:none;min-height:100vh;place-items:center;background:var(--bg-secondary);position:fixed;inset:0;z-index:200}.login-page.show{display:grid}.login-page .zyvor-logo{margin:0 auto 20px;display:flex;align-items:center;justify-content:center;gap:8px;text-decoration:none;color:var(--text)}.login-page .modal-box{width:min(400px,90vw);box-shadow:var(--shadow-lg)}.login-page h3{text-align:center;font-size:28px;font-weight:600;letter-spacing:-.02em;color:var(--text)}.login-page p{color:var(--text-secondary)}
.global-nav{position:sticky;top:0;z-index:50;background:var(--nav-blur);backdrop-filter:saturate(180%) blur(20px);-webkit-backdrop-filter:saturate(180%) blur(20px);border-bottom:1px solid var(--line-light)}.nav-inner{max-width:var(--content-max);margin:0 auto;padding:0 22px;height:48px;display:flex;align-items:center;gap:24px}.nav-brand{display:flex;align-items:center;gap:7px;text-decoration:none;color:var(--text);flex-shrink:0}.nav-brand:hover{text-decoration:none;opacity:.85}.zyvor-word{font-weight:700;font-size:15px;letter-spacing:-.03em;color:var(--text)}
.nav-links{display:flex;align-items:center;gap:4px;flex:1;justify-content:center}.nav-link{height:32px;padding:0 14px;border-radius:var(--radius-pill);display:flex;align-items:center;font-size:14px;color:var(--text-secondary);text-decoration:none;transition:.15s;white-space:nowrap}.nav-link:hover{background:var(--line-light);color:var(--text);text-decoration:none}.nav-link.active{background:var(--text);color:#fff;font-weight:500}
.nav-actions{display:flex;align-items:center;gap:8px;flex-shrink:0}.nav-icon-btn{width:34px;height:34px;border:0;background:transparent;border-radius:50%;display:grid;place-items:center;color:var(--text-secondary);position:relative}.nav-icon-btn:hover{background:var(--bg-secondary)}.bell-badge{position:absolute;right:2px;top:2px;width:15px;height:15px;border-radius:50%;display:grid;place-items:center;background:var(--danger);color:#fff;font-size:9px;font-weight:600;border:2px solid var(--bg)}.avatar{width:30px;height:30px;border-radius:50%;display:grid;place-items:center;background:var(--bg-secondary);border:1px solid var(--line);font-size:12px;font-weight:600;color:var(--text-secondary)}
.main-content{flex:1;width:100%}.page-view{display:none;padding-bottom:64px}.page-view.active{display:block}
.hero{background:var(--bg-secondary);padding:72px 22px 56px;text-align:center}.hero-inner{max-width:var(--content-max);margin:0 auto}.hero-eyebrow{font-size:14px;color:var(--text-secondary);margin-bottom:8px;font-weight:500}.hero-title{font-size:48px;font-weight:600;letter-spacing:-.025em;line-height:1.08;margin:0;color:var(--text)}.hero-sub{font-size:21px;color:var(--text-secondary);margin:14px auto 0;max-width:520px;line-height:1.4;font-weight:400}.hero-meta{display:flex;align-items:center;justify-content:center;gap:20px;margin-top:28px;flex-wrap:wrap}.hero-health{display:flex;align-items:center;gap:10px;padding:8px 16px;background:var(--bg);border-radius:var(--radius-pill);border:1px solid var(--line-light);font-size:13px;color:var(--text-secondary)}.health-ring{--health:98;width:36px;height:36px;border-radius:50%;display:grid;place-items:center;background:conic-gradient(var(--success) calc(var(--health)*1%),var(--line-light) 0);position:relative;flex-shrink:0}.health-ring:after{content:"";position:absolute;inset:4px;border-radius:50%;background:var(--bg)}.health-num{z-index:1;font-size:11px;font-weight:600;color:var(--text)}.healthy{color:#248a3d;font-weight:500}.healthy:before{content:"";display:inline-block;width:6px;height:6px;border-radius:50%;background:var(--success);margin-right:5px}
.section{padding:56px 22px}.section-alt{background:var(--bg-secondary)}.section-inner{max-width:var(--content-max);margin:0 auto}.section-head{margin-bottom:32px}.section-eyebrow{font-size:12px;text-transform:uppercase;letter-spacing:.06em;color:var(--text-tertiary);font-weight:600;margin-bottom:6px}.section-title{font-size:32px;font-weight:600;letter-spacing:-.02em;margin:0;color:var(--text)}.section-sub{font-size:17px;color:var(--text-secondary);margin-top:8px;line-height:1.4}
.metrics{display:grid;grid-template-columns:repeat(3,1fr);gap:16px}.metric{border-radius:var(--radius-lg);background:var(--bg);padding:24px 26px;border:1px solid var(--line-light);box-shadow:var(--shadow)}.section-alt .metric{background:var(--bg)}.metric-label{font-size:13px;color:var(--text-secondary);font-weight:500}.metric-value{font-size:36px;font-weight:600;letter-spacing:-.03em;margin-top:8px;color:var(--text)}.metric-trend{font-size:13px;margin-top:6px;color:var(--success);font-weight:500}.metric-trend.bad{color:var(--danger)}
.tile-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:16px}.tile{display:flex;flex-direction:column;justify-content:space-between;min-height:160px;padding:28px;border-radius:var(--radius-lg);background:var(--bg);border:1px solid var(--line-light);text-align:left;cursor:pointer;transition:transform .2s,box-shadow .2s;box-shadow:var(--shadow);color:inherit;text-decoration:none}.tile:hover{transform:scale(1.02);box-shadow:var(--shadow-lg);text-decoration:none}.section-alt .tile{background:var(--bg)}.tile-label{font-size:12px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-tertiary);font-weight:600}.tile-title{font-size:24px;font-weight:600;letter-spacing:-.02em;margin-top:8px;color:var(--text)}.tile-desc{font-size:14px;color:var(--text-secondary);margin-top:6px;line-height:1.45}.tile-link{font-size:14px;color:var(--link);margin-top:16px;font-weight:400}
.persona-row{display:flex;flex-wrap:wrap;gap:8px;justify-content:center;margin-top:24px}.persona-chip{display:inline-flex;align-items:center;gap:6px;padding:6px 12px;border-radius:var(--radius-pill);border:1px solid var(--line);background:var(--bg);font-size:12px;color:var(--text-secondary)}.persona-chip.active{border-color:rgba(52,199,89,.35);background:#f0faf3;color:#248a3d;font-weight:500}
.page-toolbar{display:flex;align-items:center;justify-content:space-between;gap:16px;margin-bottom:28px;flex-wrap:wrap}.page-toolbar h1{font-size:32px;font-weight:600;letter-spacing:-.02em;margin:0}.page-toolbar p{font-size:15px;color:var(--text-secondary);margin:4px 0 0}.toolbar-actions{display:flex;gap:8px;align-items:center;flex-wrap:wrap}
.select-btn,.refresh-btn{height:34px;border:1px solid var(--line);border-radius:var(--radius-pill);background:var(--bg);color:var(--text);font-size:13px;padding:0 16px}.refresh-btn:hover,.select-btn:hover{background:var(--bg-secondary)}
.card{border:1px solid var(--line-light);border-radius:var(--radius-lg);background:var(--bg);box-shadow:var(--shadow);overflow:hidden;margin-bottom:24px}.card-head{padding:24px 28px 12px;display:flex;align-items:flex-start;justify-content:space-between;gap:16px;flex-wrap:wrap}.card-title{font-size:21px;font-weight:600;letter-spacing:-.01em;margin:0;color:var(--text)}.card-sub{font-size:14px;color:var(--text-secondary);margin-top:4px}.card-action{border:0;background:transparent;font-size:14px;color:var(--link);padding:0;cursor:pointer}.card-action:hover{text-decoration:underline}
.topology-card .topology{position:relative;height:420px;margin:0 28px 28px;border:1px solid var(--line-light);border-radius:var(--radius);overflow:hidden;background-image:radial-gradient(var(--line) 1px,transparent 1px);background-size:20px 20px;background-color:var(--bg-secondary)}.topology-card.fullscreen{position:fixed;inset:20px;z-index:90;height:auto!important;box-shadow:var(--shadow-lg)}.topology-card.fullscreen .topology{height:calc(100vh - 120px)!important;margin:0 28px 28px}.topo-canvas{position:absolute;inset:0;transform-origin:50% 50%;transition:transform .18s ease}.topo-toolbar{position:absolute;left:12px;top:12px;z-index:3;display:flex;border:1px solid var(--line);border-radius:var(--radius-sm);overflow:hidden;background:var(--bg);box-shadow:var(--shadow)}.topo-toolbar button{width:36px;height:32px;border:0;border-right:1px solid var(--line-light);background:none;font-size:13px;color:var(--text)}.topo-toolbar button:last-child{border:0}.topo-lines{position:absolute;inset:0;width:100%;height:100%}.topo-lines line{stroke:var(--line);stroke-width:1.8}.topo-node{position:absolute;z-index:2;display:flex;align-items:center;gap:8px}.topo-shape{width:40px;height:40px;border-radius:var(--radius-sm);border:1px solid var(--line);background:var(--bg);display:grid;place-items:center;color:var(--accent);font-size:18px;box-shadow:var(--shadow)}.topo-node.green .topo-shape{border-color:rgba(52,199,89,.4);background:#f0faf3;color:#248a3d}.topo-node.orange .topo-shape{border-color:rgba(255,149,0,.4);background:#fff8f0;color:#c93400}.topo-node.purple .topo-shape{border-color:rgba(175,82,222,.4);background:#faf5fc;color:#8944ab}.topo-text strong{display:block;font-size:13px;font-weight:600}.topo-text span{display:block;color:var(--text-secondary);font-size:11px;margin-top:2px}.tn-dc{left:49%;top:21px}.tn-c1{left:25%;top:97px}.tn-c2{right:18%;top:97px}.tn-h1{left:8%;top:187px}.tn-h2{left:35%;top:187px}.tn-h3{right:16%;top:187px}.tn-ds1{left:9%;bottom:9px}.tn-ds2{left:40%;bottom:9px}.tn-ds3{right:11%;bottom:9px}
.split-grid{display:grid;grid-template-columns:1fr 1fr;gap:24px}.activity-wrap{display:grid;grid-template-columns:180px 1fr;align-items:center;padding:24px 28px 28px;gap:24px}.donut{--parts:conic-gradient(#e8e8ed 0 100%);width:160px;height:160px;border-radius:50%;background:var(--parts);position:relative;display:grid;place-items:center;margin:0 auto}.donut:after{content:"";position:absolute;inset:28px;border-radius:50%;background:var(--bg);border:1px solid var(--line-light)}.donut-center{z-index:1;text-align:center;font-size:22px;font-weight:600}.donut-center small{font-size:12px;color:var(--text-secondary);display:block;font-weight:400}.legend{display:grid;gap:14px}.legend-row{display:grid;grid-template-columns:10px 1fr auto;gap:10px;align-items:center;font-size:14px}.legend-dot{width:10px;height:10px;border-radius:50%}.legend-row span:last-child{color:var(--text-secondary)}.activity-empty,.request-empty{padding:48px 28px;color:var(--text-tertiary);text-align:center;font-size:14px}
.request-list{padding:0 28px 20px}.request-row{height:38px;display:grid;grid-template-columns:52px minmax(0,1fr) 56px 56px;gap:10px;align-items:center;border-top:1px solid var(--line-light);font-size:13px;color:var(--text-secondary)}.request-row:hover{background:var(--bg-secondary)}.method{font-size:11px;padding:4px 8px;border-radius:var(--radius-pill);text-align:center;background:var(--bg-secondary);color:var(--text);font-weight:500}.method.get{background:#e8f2ff;color:var(--link)}.request-path{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--text)}.request-status{color:var(--success);text-align:right;font-weight:500}.request-status.err{color:var(--danger)}.request-ms{text-align:right;color:var(--text-tertiary)}
.vm-head-tools{display:flex;gap:8px;flex-wrap:wrap}.table-search{width:180px;height:34px;border:1px solid var(--line);background:var(--bg-secondary);border-radius:var(--radius-pill);display:flex;align-items:center;padding:0 14px;gap:8px;color:var(--text-tertiary)}.table-search input{width:100%;border:0;background:none;outline:0;font-size:13px;color:var(--text)}.mini-select{height:34px;border:1px solid var(--line);background:var(--bg);border-radius:var(--radius-pill);font-size:13px;padding:0 14px;color:var(--text)}.export-btn{height:34px;border:0;background:var(--accent);border-radius:var(--radius-pill);padding:0 18px;font-size:13px;color:#fff;font-weight:500}.export-btn:hover{background:var(--accent-hover)}.table-shell{padding:0 28px 20px}.vm-table{width:100%;border-collapse:collapse;table-layout:fixed}.vm-table thead{background:var(--bg-secondary)}.vm-table th{height:40px;text-align:left;font-size:11px;color:var(--text-secondary);font-weight:600;border-top:1px solid var(--line-light);border-bottom:1px solid var(--line-light);padding:0 14px;text-transform:uppercase;letter-spacing:.04em}.vm-table td{height:48px;border-bottom:1px solid var(--line-light);padding:0 14px;font-size:14px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.vm-table tr:hover td{background:var(--bg-secondary)}.vm-state{display:inline-flex;align-items:center;gap:5px}.vm-state:before{content:"";width:6px;height:6px;border-radius:50%;background:var(--text-tertiary)}.vm-state.on:before{background:var(--success)}.vm-state.suspended:before{border-radius:1px;background:var(--warning)}.vm-fixture{display:inline-flex;padding:3px 10px;border-radius:var(--radius-pill);font-size:11px;border:1px solid var(--line);color:var(--text-secondary);background:var(--bg-secondary)}.vm-fixture.real{border-color:rgba(52,199,89,.35);background:#f0faf3;color:#248a3d}.vm-actions{display:flex;justify-content:flex-end;gap:6px}.icon-btn{width:28px;height:28px;border:1px solid var(--line);background:var(--bg);border-radius:50%;display:grid;place-items:center;color:var(--text-secondary);font-size:13px}.icon-btn:hover{background:var(--bg-secondary)}.pager{height:44px;display:flex;align-items:center;justify-content:space-between;color:var(--text-secondary);font-size:13px;padding:0 28px 16px}.pages{display:flex;gap:4px}.page-btn{min-width:28px;height:28px;border:0;background:transparent;color:var(--text-secondary);font-size:13px;border-radius:50%}.page-btn.active{background:var(--accent);color:#fff}
.scenario-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:16px;padding:8px 28px 28px}.scenario{min-height:140px;border-radius:var(--radius);text-align:left;padding:24px 24px 24px 28px;border:1px solid var(--line-light);background:var(--bg);position:relative;cursor:pointer;box-shadow:var(--shadow);transition:transform .15s}.scenario:hover{transform:translateY(-2px)}.scenario:before{content:"";position:absolute;left:0;top:0;bottom:0;width:4px;background:var(--success);border-radius:4px 0 0 4px}.scenario.slow:before{background:var(--warning)}.scenario.flaky:before{background:var(--danger)}.scenario.resume:before{background:var(--accent)}.scenario strong{font-size:17px;font-weight:600;display:block}.scenario p{font-size:14px;color:var(--text-secondary);margin:8px 0 0;line-height:1.4;max-width:85%}.scenario-run{position:absolute;right:20px;bottom:20px;width:36px;height:36px;border-radius:50%;display:grid;place-items:center;border:0;color:#fff;background:var(--accent);font-size:13px}.scenario.slow .scenario-run{background:var(--warning)}.scenario.flaky .scenario-run{background:var(--danger)}.scenario.active{box-shadow:0 0 0 2px var(--accent)}.view-scenarios{display:block;width:calc(100% - 56px);margin:0 28px 28px;height:44px;border:1px solid var(--line);border-radius:var(--radius-pill);background:var(--bg);color:var(--link);font-size:14px}.view-scenarios:hover{background:var(--bg-secondary)}
.site-footer{background:var(--bg-secondary);border-top:1px solid var(--line-light);padding:32px 22px}.footer-inner{max-width:var(--content-max);margin:0 auto;display:grid;grid-template-columns:repeat(3,1fr) auto;gap:24px;align-items:start;font-size:13px}.footer-label{font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--text-tertiary);font-weight:600;margin-bottom:4px}.footer-value{color:var(--text);font-weight:500}.footer-value small{display:block;color:var(--text-secondary);font-weight:400;margin-top:2px;font-size:12px}.online{color:#248a3d;font-weight:500}.copyright{text-align:right;color:var(--text-secondary);font-size:12px}.copyright small{display:block;color:var(--text-tertiary);margin-top:4px}
.drawer{position:fixed;right:-460px;top:0;width:440px;height:100vh;background:var(--bg);border-left:1px solid var(--line);z-index:80;box-shadow:var(--shadow-lg);transition:right .25s ease;display:flex;flex-direction:column}.drawer.open{right:0}.drawer-head{height:68px;padding:0 24px;border-bottom:1px solid var(--line-light);display:flex;align-items:center;justify-content:space-between}.drawer-head h3{font-size:21px;font-weight:600;margin:0}.drawer-head p{font-size:13px;color:var(--text-secondary);margin:2px 0 0}.close{width:32px;height:32px;border:1px solid var(--line);border-radius:50%;background:var(--bg);color:var(--text-secondary)}.drawer-body{padding:24px;overflow:auto;flex:1}.fault-summary{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-bottom:20px}.fault-chip{border:1px solid var(--line);background:var(--bg-secondary);border-radius:var(--radius-sm);padding:12px}.fault-chip small{display:block;color:var(--text-secondary);font-size:11px}.fault-chip strong{display:block;font-size:16px;margin-top:4px;font-weight:600}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:14px}.field label{display:block;color:var(--text-secondary);font-size:12px;margin-bottom:6px;font-weight:500}.field input,.field select{width:100%;height:40px;border:1px solid var(--line);background:var(--bg-secondary);border-radius:var(--radius-sm);padding:0 12px;outline:none;font-size:14px}.field input:focus{border-color:var(--accent);box-shadow:0 0 0 3px rgba(0,113,227,.15)}.field.full{grid-column:1/-1}.drawer-actions{display:flex;gap:10px;margin-top:20px;flex-wrap:wrap}.action-btn{height:40px;border-radius:var(--radius-pill);border:1px solid var(--line);background:var(--bg);padding:0 18px;font-size:14px}.action-btn.primary{background:var(--accent);border-color:var(--accent);color:#fff;font-weight:500}.lock-note{margin-top:16px;padding:12px;border:1px solid var(--line);border-radius:var(--radius-sm);background:var(--bg-secondary);color:var(--text-secondary);font-size:13px;line-height:1.45}
.overlay{position:fixed;inset:0;background:rgba(0,0,0,.28);z-index:70;opacity:0;pointer-events:none;transition:.2s}.overlay.open{opacity:1;pointer-events:auto}
.modal{position:fixed;inset:0;z-index:100;background:rgba(0,0,0,.28);display:none;place-items:center;backdrop-filter:blur(8px)}.modal.open{display:grid}.modal-box{width:min(420px,90vw);border:1px solid var(--line);border-radius:var(--radius-lg);background:var(--bg);padding:28px;box-shadow:var(--shadow-lg)}.modal-box h3{margin:0;font-size:21px;font-weight:600}.modal-box p{color:var(--text-secondary);font-size:14px}.modal-box input,.modal-box select{width:100%;height:40px;border:1px solid var(--line);background:var(--bg-secondary);border-radius:var(--radius-sm);padding:0 12px;outline:none;font-size:14px}.modal-actions{display:flex;justify-content:flex-end;gap:10px;margin-top:16px}.toast{position:fixed;right:22px;bottom:22px;z-index:110;width:320px;border:1px solid var(--line);border-radius:var(--radius);background:var(--bg);padding:14px 16px;box-shadow:var(--shadow-lg);transform:translateY(20px);opacity:0;pointer-events:none;transition:.2s}.toast.show{transform:none;opacity:1}.toast strong{font-size:14px;font-weight:600}.toast span{display:block;color:var(--text-secondary);font-size:12px;margin-top:2px}.browse-list{border:1px solid var(--line);border-radius:var(--radius-sm);background:var(--bg)}
.visually-hidden{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);border:0}
@media(max-width:900px){.metrics{grid-template-columns:repeat(2,1fr)}.tile-grid,.split-grid,.scenario-grid{grid-template-columns:1fr}.hero-title{font-size:36px}.nav-links{display:none}.footer-inner{grid-template-columns:1fr 1fr}.copyright{display:none}}
@media(max-width:600px){.metrics{grid-template-columns:1fr}.hero{padding:48px 16px 40px}.section{padding:40px 16px}.nav-inner{padding:0 16px}.footer-inner{grid-template-columns:1fr}.drawer{width:100%;right:-100%}}
</style>
</head>
<body>
<noscript><div style="padding:40px;text-align:center;font-family:-apple-system,sans-serif"><h1>Chimera Command Center</h1><p>JavaScript is required. Enable it and reload <a href="/__chimera/">/__chimera/</a>.</p></div></noscript>
<div class="shell" id="appRoot">
<nav class="global-nav">
  <div class="nav-inner">
    <a class="nav-brand" href="https://zyvor.dev" target="_blank" rel="noopener" aria-label="Zyvor home"><svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true"><path d="M2 2h14L6.6 16H16" fill="none" stroke="#ff5a15" stroke-width="2.6" stroke-linejoin="round" stroke-linecap="round"/></svg><span class="zyvor-word">zyvor</span></a>
    <div class="nav-links">
      <a class="nav-link active" href="#home" data-page="home">Home</a>
      <a class="nav-link" href="#infrastructure" data-page="infrastructure">Infrastructure</a>
      <a class="nav-link" href="#inventory" data-page="inventory">Inventory</a>
      <a class="nav-link" href="#telemetry" data-page="telemetry">Telemetry</a>
      <a class="nav-link" href="#lab" data-page="lab">Lab</a>
    </div>
    <div class="nav-actions">
      <button class="nav-icon-btn" id="faultTop" title="Fault Studio">ϟ</button>
      <button class="nav-icon-btn" id="bellBtn" title="Recent errors">♧<span class="bell-badge" id="alertBadge">0</span></button>
      <button class="nav-icon-btn" id="gearBtn" title="Settings">⚙</button>
      <button class="nav-icon-btn" id="userMenu" title="Log out"><div class="avatar">AD</div></button>
    </div>
  </div>
</nav>

<main class="main-content">

<section class="page-view active" id="page-home">
  <div class="hero">
    <div class="hero-inner">
      <div class="hero-eyebrow"><span id="footerPersona">vSphere</span> · Simulation Engine</div>
      <h1 class="hero-title">Your infrastructure lab.<br/>Always on.</h1>
      <p class="hero-sub">Run migration tests, inject faults, and watch live traffic — without provisioning real hardware.</p>
      <div class="hero-meta">
        <div class="hero-health"><div class="health-ring" id="healthRing"><div class="health-num"><span id="healthPct">98</span></div></div><span class="healthy" id="healthLabel">Healthy</span><span>· Uptime <strong id="sideUptime">0m</strong></span></div>
        <div class="hero-health"><span id="footerEndpoint">loading…</span></div>
      </div>
      <div class="persona-row" id="personaList">
        <span class="persona-chip active">▣ vSphere</span>
        <span class="persona-chip">Λ Nutanix</span>
        <span class="persona-chip">✕ Proxmox</span>
        <span class="persona-chip">▦ OpenStack</span>
        <span class="persona-chip">⊞ Hyper-V</span>
      </div>
    </div>
  </div>
  <div class="section">
    <div class="section-inner">
      <div class="section-head"><div class="section-eyebrow">Live metrics</div><h2 class="section-title">Gateway telemetry</h2><p class="section-sub">Real-time counters from your simulated estate.</p></div>
      <div class="metrics">
        <div class="metric"><div class="metric-label">Requests</div><div class="metric-value" id="kRequests">0</div><div class="metric-trend" id="tRequests">↑ live</div></div>
        <div class="metric"><div class="metric-label">Error Rate</div><div class="metric-value" id="kError">0.00%</div><div class="metric-trend" id="tError">↓ clean</div></div>
        <div class="metric"><div class="metric-label">Active Sessions</div><div class="metric-value" id="kSessions">0</div><div class="metric-trend" id="tSessions">↑ listening</div></div>
        <div class="metric"><div class="metric-label">Data Transfer</div><div class="metric-value" id="kBytes">0 B</div><div class="metric-trend" id="tBytes">↑ NFC + API</div></div>
        <div class="metric"><div class="metric-label">Exports</div><div class="metric-value" id="kExports">0</div><div class="metric-trend" id="tExports">↑ NFC leases</div></div>
        <div class="metric"><div class="metric-label">Avg Response</div><div class="metric-value" id="kLatency">0 ms</div><div class="metric-trend" id="tLatency">↓ realtime</div></div>
      </div>
    </div>
  </div>
  <div class="section section-alt">
    <div class="section-inner">
      <div class="section-head"><div class="section-eyebrow">Explore</div><h2 class="section-title">Everything in one engine.</h2></div>
      <div class="tile-grid">
        <a class="tile" href="#infrastructure" data-page="infrastructure"><div><div class="tile-label">Infrastructure</div><div class="tile-title">Topology</div><div class="tile-desc">Datacenter → cluster → host → datastore map of your simulated estate.</div></div><div class="tile-link">View topology →</div></a>
        <a class="tile" href="#inventory" data-page="inventory"><div><div class="tile-label">Inventory</div><div class="tile-title">Virtual machines</div><div class="tile-desc"><span id="sideVmCount">0</span> discoverable VMs and VMDK fixture library.</div></div><div class="tile-link">Browse inventory →</div></a>
        <a class="tile" href="#telemetry" data-page="telemetry"><div><div class="tile-label">Telemetry</div><div class="tile-title">Live traffic</div><div class="tile-desc">Request feed and operation-class breakdown from the gateway.</div></div><div class="tile-link">Watch traffic →</div></a>
        <a class="tile" href="#lab" data-page="lab"><div><div class="tile-label">Lab</div><div class="tile-title">Scenarios &amp; faults</div><div class="tile-desc">One-click failure environments and Fault Studio controls.</div></div><div class="tile-link">Open lab →</div></a>
      </div>
    </div>
  </div>
</section>

<section class="page-view" id="page-infrastructure">
  <div class="section" style="padding-top:48px">
    <div class="section-inner">
      <div class="page-toolbar"><div><h1>Infrastructure</h1><p id="topologySub">Live simulated estate</p></div><div class="toolbar-actions"><button class="refresh-btn" id="refreshBtn">↻ Refresh</button></div></div>
      <div class="card topology-card" id="topology">
        <div class="card-head"><div><h2 class="card-title">Topology</h2><div class="card-sub">Visual map of datacenters, clusters, hosts, and datastores</div></div><button class="card-action" id="viewFullTopology">Full screen</button></div>
        <div class="topology"><div class="topo-toolbar"><button id="fitTopology">Fit</button><button id="zoomTopology">＋</button></div>
          <div class="topo-canvas">
          <svg class="topo-lines" viewBox="0 0 640 318" preserveAspectRatio="none"><line x1="330" y1="59" x2="185" y2="122"/><line x1="330" y1="59" x2="480" y2="122"/><line x1="185" y1="142" x2="85" y2="212"/><line x1="185" y1="142" x2="255" y2="212"/><line x1="480" y1="142" x2="515" y2="212"/><line x1="85" y1="230" x2="90" y2="282"/><line x1="255" y1="230" x2="285" y2="282"/><line x1="515" y1="230" x2="540" y2="282"/></svg>
          <div class="topo-node green tn-dc"><div class="topo-shape">⬡</div><div class="topo-text"><strong>DC0</strong><span>Datacenter</span></div></div>
          <div class="topo-node tn-c1"><div class="topo-shape">◇</div><div class="topo-text"><strong>Cluster0</strong><span>Cluster</span></div></div>
          <div class="topo-node tn-c2"><div class="topo-shape">◇</div><div class="topo-text"><strong>Cluster1</strong><span>Cluster</span></div></div>
          <div class="topo-node orange tn-h1"><div class="topo-shape">▥</div><div class="topo-text"><strong>Host0</strong><span>192.168.1.10</span></div></div>
          <div class="topo-node orange tn-h2"><div class="topo-shape">▥</div><div class="topo-text"><strong>Host1</strong><span>192.168.1.11</span></div></div>
          <div class="topo-node orange tn-h3"><div class="topo-shape">▥</div><div class="topo-text"><strong>Host2</strong><span>192.168.1.12</span></div></div>
          <div class="topo-node purple tn-ds1"><div class="topo-shape">◉</div><div class="topo-text"><strong>Datastore0</strong><span id="dsMeta0">fixture</span></div></div>
          <div class="topo-node purple tn-ds2"><div class="topo-shape">◉</div><div class="topo-text"><strong>Datastore1</strong><span id="dsMeta1">fixture</span></div></div>
          <div class="topo-node purple tn-ds3"><div class="topo-shape">◉</div><div class="topo-text"><strong>Datastore2</strong><span id="dsMeta2">fixture</span></div></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<section class="page-view" id="page-inventory">
  <div class="section" style="padding-top:48px">
    <div class="section-inner">
      <div class="page-toolbar"><div><h1>Inventory</h1><p>Discoverable vSphere virtual machines and fixture disks</p></div></div>
      <div class="card" id="inventory">
        <div class="card-head"><div><h2 class="card-title">Virtual Machines (<span id="vmTotal">0</span>)</h2></div>
          <div class="vm-head-tools"><div class="table-search">⌕<input id="vmSearch" placeholder="Search VMs…" /></div><select id="powerFilter" class="mini-select"><option value="all">All states</option><option value="poweredOn">Powered On</option><option value="poweredOff">Powered Off</option></select><button class="export-btn" id="copyVmBtn">Export VM</button></div>
        </div>
        <div class="table-shell"><table class="vm-table"><thead><tr><th style="width:20%">Name</th><th style="width:14%">State</th><th style="width:8%">CPU</th><th style="width:9%">Memory</th><th style="width:8%">Disks</th><th style="width:14%">Datastore</th><th style="width:13%">Fixture</th><th style="width:14%;text-align:right">Actions</th></tr></thead><tbody id="vmRows"></tbody></table><div class="pager"><span id="pagerInfo">Showing 0 VMs</span><div class="pages" id="pages"></div></div></div>
      </div>
      <div class="card" id="vmdks">
        <div class="card-head"><div><h2 class="card-title">VMDK Library (<span id="vmdkTotal">0</span>)</h2><div class="card-sub" id="vmdkSub">No fixture_vmdk_dir configured</div></div><button class="export-btn" id="openUploadBtn">Upload VMDK</button></div>
        <div class="table-shell"><table class="vm-table"><thead><tr><th style="width:34%">File</th><th style="width:16%">Size</th><th style="width:24%">Assigned VM</th><th style="width:26%">Method</th></tr></thead><tbody id="vmdkRows"></tbody></table></div>
      </div>
    </div>
  </div>
</section>

<section class="page-view" id="page-telemetry">
  <div class="section" style="padding-top:48px">
    <div class="section-inner">
      <div class="page-toolbar"><div><h1>Telemetry</h1><p>Traffic breakdown and live request feed</p></div></div>
      <div class="split-grid">
        <div class="card" id="activity"><div class="card-head"><div><h2 class="card-title">Top Activity</h2><div class="card-sub">By operation class</div></div></div><div class="activity-wrap"><div class="donut" id="donut"><div class="donut-center"><span id="donutTotal">0</span><small>Total</small></div></div><div class="legend" id="activityLegend"><div class="activity-empty">No traffic yet.</div></div></div></div>
        <div class="card" id="requests"><div class="card-head"><div><h2 class="card-title">Live Requests</h2><div class="card-sub">API and transfer activity</div></div><button class="card-action" id="viewAllRequests">View all</button></div><div class="request-list" id="requestList"><div class="request-empty">Waiting for client traffic…</div></div></div>
      </div>
    </div>
  </div>
</section>

<section class="page-view" id="page-lab">
  <div class="section" style="padding-top:48px">
    <div class="section-inner">
      <div class="page-toolbar"><div><h1>Lab</h1><p>Deterministic scenarios and fault injection</p></div><button class="export-btn" id="openFaultsBottom">Fault Studio</button></div>
      <div class="card" id="scenarios">
        <div class="card-head"><div><h2 class="card-title">Scenario Launcher</h2><div class="card-sub">One-click failure environments</div></div></div>
        <div class="scenario-grid">
          <button class="scenario clean" data-scenario="clean"><strong>Clean Environment</strong><p>Reset all faults and counters</p><span class="scenario-run">◯</span></button>
          <button class="scenario slow" data-scenario="slow"><strong>Slow Fabric</strong><p>High latency and low bandwidth</p><span class="scenario-run">▷</span></button>
          <button class="scenario flaky" data-scenario="flaky"><strong>Flaky API</strong><p>Random failures and timeouts</p><span class="scenario-run">◉</span></button>
          <button class="scenario resume" data-scenario="resume"><strong>Resume Export</strong><p>Drop connection mid-transfer</p><span class="scenario-run">▷</span></button>
        </div>
      </div>
    </div>
  </div>
</section>

<footer class="site-footer">
  <div class="footer-inner">
    <div><div class="footer-label">Persona</div><div class="footer-value"><span class="online">Active</span> · vSphere</div></div>
    <div><div class="footer-label">Administrator</div><div class="footer-value" id="footerUser">administrator@vsphere.local</div></div>
    <div><div class="footer-label">System time</div><div class="footer-value" id="systemTime">--</div></div>
    <div class="copyright">© 2026 Zyvor Chimera<small>Version <span id="sideVersion">dev</span></small></div>
  </div>
</footer>
</main>
</div>

<div class="overlay" id="overlay"></div>
<div class="overlay" id="topoOverlay" style="z-index:85"></div>
<aside class="drawer" id="faultDrawer"><div class="drawer-head"><div><h3>Fault Studio</h3><p>Shape latency, errors, drops and bandwidth.</p></div><button class="close" id="closeDrawer">×</button></div><div class="drawer-body"><div class="fault-summary"><div class="fault-chip"><small>Latency</small><strong id="faultLatencySummary">0 ms</strong></div><div class="fault-chip"><small>Bandwidth</small><strong id="faultBandwidthSummary">∞</strong></div><div class="fault-chip"><small>Status</small><strong id="faultStatusSummary">503</strong></div></div><div class="form-grid">
<div class="field"><label>Latency (ms)</label><input id="fLatency" type="number" min="0" value="0" /></div><div class="field"><label>Failure HTTP status</label><input id="fStatus" type="number" min="400" max="599" value="503" /></div>
<div class="field"><label>Fail next requests</label><input id="fFail" type="number" min="0" value="0" /></div><div class="field"><label>Fail next NFC requests</label><input id="fNfcFail" type="number" min="0" value="0" /></div>
<div class="field"><label>Drop next NFC streams</label><input id="fDrop" type="number" min="0" value="0" /></div><div class="field"><label>Drop after (MiB)</label><input id="fDropAfter" type="number" min="0" step="0.25" value="0" /></div>
<div class="field full"><label>Bandwidth cap (MiB/s, 0 = unlimited)</label><input id="fBandwidth" type="number" min="0" step="0.25" value="0" /></div></div><div class="drawer-actions"><button class="action-btn primary" id="applyFaults">Apply policy</button><button class="action-btn" id="resetFaults">Reset</button><button class="action-btn" id="unlockBtn">Unlock</button></div><div class="lock-note" id="lockNote">🔒 Administrative actions require logging in. The token stays in this browser tab.</div></div></aside>
<aside class="drawer" id="settingsDrawer"><div class="drawer-head"><div><h3>Settings</h3><p>Server info and admin login.</p></div><button class="close" id="closeSettingsDrawer">×</button></div><div class="drawer-body">
<div class="fault-summary" style="grid-template-columns:1fr"><div class="fault-chip"><small>Listen address</small><strong id="settingsListen">—</strong></div></div>
<div class="lock-note" style="margin-bottom:16px">Set <b>CHIMERA_LISTEN</b> and restart to change the listen address/port.</div>
<div class="form-grid">
<div class="field full"><label>New admin username</label><input id="settingsUser" type="text" autocomplete="off" placeholder="admin" /></div>
<div class="field full"><label>New admin password</label><input id="settingsPass" type="password" autocomplete="new-password" placeholder="••••••" /></div>
</div>
<div class="drawer-actions"><button class="action-btn primary" id="saveCredentials">Save login</button></div>
</div></aside>
<div class="login-page show" id="loginPage"><div class="modal-box"><a class="zyvor-logo" href="https://zyvor.dev" target="_blank" rel="noopener" aria-label="Zyvor home"><svg width="22" height="22" viewBox="0 0 18 18" aria-hidden="true"><path d="M2 2h14L6.6 16H16" fill="none" stroke="#ff5a15" stroke-width="2.6" stroke-linejoin="round" stroke-linecap="round"/></svg><span class="zyvor-word" style="font-size:18px">zyvor</span></a><h3>Log in to Chimera</h3><p>Default is admin / admin unless changed in Settings.</p><input id="authUser" type="text" autocomplete="username" placeholder="Username" style="margin-bottom:8px" /><input id="authPass" type="password" autocomplete="current-password" placeholder="Password" /><div class="modal-actions" style="justify-content:center"><button class="action-btn primary" id="authSave">Log in</button></div></div></div>
<div class="modal" id="uploadModal"><div class="modal-box" style="width:min(560px,92vw)"><h3>Add a VMDK</h3><p>Upload a new file, or pick one already staged on the host.</p><div class="field full" style="margin:14px 0"><label>Upload a file</label><input id="uploadFile" type="file" accept=".vmdk" /></div><div class="field full"><label>Assign to VM</label><select id="uploadVmSelect"><option value="">Auto (name-match / round-robin)</option></select></div><div class="modal-actions"><button class="action-btn" id="uploadCancel">Cancel</button><button class="action-btn primary" id="uploadSubmit">Upload</button></div><div class="field full" style="margin-top:16px;border-top:1px solid var(--line-light);padding-top:14px"><label>Or pick a file on the host</label><div style="display:flex;align-items:center;gap:6px;margin:6px 0"><select id="browseRootSelect" class="mini-select"></select><button class="icon-btn" id="browseUpBtn" title="Up">↑</button><span id="browsePathLabel" style="font-size:11px;color:var(--text-secondary);flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">/</span></div><div id="browseList" class="browse-list" style="max-height:190px;overflow:auto"></div></div></div></div>
<div class="toast" id="toast"><strong id="toastTitle">Done</strong><span id="toastBody"></span></div>
<input type="search" id="globalSearch" class="visually-hidden" tabindex="-1" aria-hidden="true" />

<script>
(function(){
  var q=function(s){return document.querySelector(s)};var qa=function(s){return Array.prototype.slice.call(document.querySelectorAll(s))};
  var base='/__chimera',token=sessionStorage.getItem('chimeraToken')||'',bootstrap=null,inventory={virtual_machines:[]},telemetry=null,state=null,vmdks={files:[],total:0,directory:'',roots:[]};
  var browseRoot=0,browsePath='',vmPage=1,pageSize=8,lastTelemetry=null,colors=['#0071e3','#34c759','#ff9500','#ff3b30','#af52de','#5856d6'];
  var pages=['home','infrastructure','inventory','telemetry','lab'];
  function esc(v){return String(v==null?'':v).replace(/[&<>"']/g,function(m){return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#039;'}[m]})}
  function fmtBytes(n){n=Number(n||0);if(n<1024)return n+' B';if(n<1048576)return (n/1024).toFixed(1)+' KB';if(n<1073741824)return (n/1048576).toFixed(1)+' MB';return (n/1073741824).toFixed(1)+' GB'}
  function fmtUptime(s){s=Number(s||0);if(s<60)return Math.floor(s)+'s';if(s<3600)return Math.floor(s/60)+'m';if(s<86400)return Math.floor(s/3600)+'h '+Math.floor((s%3600)/60)+'m';return Math.floor(s/86400)+'d '+Math.floor((s%86400)/3600)+'h'}
  function fmtBps(v){v=Number(v||0);return v<=0?'∞':(v/1048576).toFixed(v%1048576?1:0)+' MiB/s'}
  function fixtureLabel(m){return {'name-match':'Matched','round-robin':'Round-robin','shared-file':'Shared','manual':'Manual','generated':'Synthetic'}[m]||'—'}
  function fixtureClass(m){return (m==='name-match'||m==='round-robin'||m==='shared-file'||m==='manual')?'real':'synthetic'}
  function toast(a,b){q('#toastTitle').textContent=a;q('#toastBody').textContent=b||'';q('#toast').classList.add('show');setTimeout(function(){q('#toast').classList.remove('show')},2200)}
  async function api(path,opts){opts=opts||{};var headers=opts.headers||{};if(token)headers.Authorization='Bearer '+token;if(opts.body&&typeof opts.body==='string'&&!headers['Content-Type'])headers['Content-Type']='application/json';var r=await fetch(base+path,Object.assign({},opts,{headers:headers}));if(r.status===401)throw new Error('Admin token required');if(!r.ok)throw new Error((await r.text())||('HTTP '+r.status));return r.json()}
  var activeDrawer='faultDrawer';
  function openDrawer(id){activeDrawer=id||'faultDrawer';q('#'+activeDrawer).classList.add('open');q('#overlay').classList.add('open')}
  function closeDrawer(){q('#'+activeDrawer).classList.remove('open');q('#overlay').classList.remove('open')}
  function showApp(){q('#appRoot').classList.add('show');q('#loginPage').classList.remove('show')}
  function showLoginPage(){q('#appRoot').classList.remove('show');q('#loginPage').classList.add('show');setTimeout(function(){q('#authUser').focus()},80)}
  function showPage(id){if(pages.indexOf(id)<0)id='home';qa('.page-view').forEach(function(p){p.classList.toggle('active',p.id==='page-'+id)});qa('.nav-link').forEach(function(n){n.classList.toggle('active',n.dataset.page===id)});if(location.hash!=='#'+id)location.hash=id;window.scrollTo(0,0)}
  function lock(){token='';sessionStorage.removeItem('chimeraToken');q('#unlockBtn').textContent='Unlock';q('#lockNote').innerHTML='🔒 Administrative actions require logging in. The token stays in this browser tab.';showLoginPage()}
  async function login(username,password){username=(username||'').trim();if(!username||!password)return;try{var r=await fetch(base+'/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:username,password:password})});if(!r.ok)throw new Error((await r.text())||'Invalid username or password');var data=await r.json();token=data.token;state=await api('/state');sessionStorage.setItem('chimeraToken',token);q('#authPass').value='';q('#unlockBtn').textContent='Lock';q('#lockNote').textContent='✓ Administrative controls unlocked for this tab.';showApp();var pg=(location.hash||'#home').slice(1);showPage(pages.indexOf(pg)>=0?pg:'home');await boot();toast('Logged in','Scenarios and fault injection are enabled.')}catch(e){lock();toast('Login failed',e.message)}}
  function renderBootstrap(){if(!bootstrap)return;q('#sideVersion').textContent=bootstrap.version;q('#sideVmCount').textContent=bootstrap.vms;q('#vmTotal').textContent=bootstrap.vms;q('#footerPersona').textContent=bootstrap.persona;q('#footerEndpoint').textContent=bootstrap.endpoint;q('#footerUser').textContent=bootstrap.username;q('#topologySub').textContent=bootstrap.datacenters+' datacenter · '+bootstrap.clusters+' clusters · '+bootstrap.hosts+' hosts';var f=bootstrap.fixture_size_mb+' MiB fixture';q('#dsMeta0').textContent=f;q('#dsMeta1').textContent=f;q('#dsMeta2').textContent=f;q('#settingsListen').textContent=bootstrap.listen||'—'}
  function filteredVMs(){var s=(q('#vmSearch').value||'').toLowerCase(),p=q('#powerFilter').value;return (inventory.virtual_machines||[]).filter(function(v){return (!s||v.name.toLowerCase().indexOf(s)>=0)&&(p==='all'||v.state===p)})}
  function renderVMs(){var all=filteredVMs(),pages=Math.max(1,Math.ceil(all.length/pageSize));if(vmPage>pages)vmPage=pages;var start=(vmPage-1)*pageSize,rows=all.slice(start,start+pageSize);q('#vmRows').innerHTML=rows.map(function(v){var st=v.state==='poweredOn'?'on':(v.state==='suspended'?'suspended':'');return '<tr><td>'+esc(v.name)+'</td><td><span class="vm-state '+st+'">'+esc(v.state)+'</span></td><td>'+v.cpu+' vCPU</td><td>'+v.memory_gb+' GB</td><td>1 ('+v.disk_gb+' GB)</td><td>'+esc(v.datastore)+'</td><td><span class="vm-fixture '+fixtureClass(v.fixture_source)+'" title="'+esc(v.fixture_file||'')+'">'+esc(fixtureLabel(v.fixture_source))+'</span></td><td><div class="vm-actions"><button class="icon-btn vm-copy" data-vm="'+esc(v.name)+'">⇩</button></div></td></tr>'}).join('')||'<tr><td colspan="8" class="empty-msg">No matching virtual machines.</td></tr>';q('#pagerInfo').textContent=all.length?'Showing '+(start+1)+'–'+Math.min(start+pageSize,all.length)+' of '+all.length:'Showing 0 VMs';var html='';for(var i=1;i<=pages&&i<=7;i++)html+='<button class="page-btn '+(i===vmPage?'active':'')+'" data-page="'+i+'">'+i+'</button>';q('#pages').innerHTML=html;qa('.page-btn').forEach(function(b){b.onclick=function(){vmPage=+b.dataset.page;renderVMs()}});qa('.vm-copy').forEach(function(b){b.onclick=function(){navigator.clipboard&&navigator.clipboard.writeText(b.dataset.vm);toast('VM selected',b.dataset.vm)}})}
  function renderVmdks(){q('#vmdkTotal').textContent=vmdks.total||0;q('#vmdkSub').textContent=vmdks.directory?vmdks.directory:'No fixture_vmdk_dir configured';q('#vmdkRows').innerHTML=(vmdks.files||[]).map(function(f){return '<tr><td>'+esc(f.file_name)+'</td><td>'+fmtBytes(f.size_bytes)+'</td><td>'+(f.vm_name?esc(f.vm_name):'—')+'</td><td>'+esc(fixtureLabel(f.method))+'</td></tr>'}).join('')||'<tr><td colspan="4" class="empty-msg">No VMDK directory configured.</td></tr>'}
  async function loadBrowse(){q('#browsePathLabel').textContent='/'+browsePath;try{var data=await api('/api/vmdks/browse?root='+browseRoot+'&path='+encodeURIComponent(browsePath));renderBrowse(data)}catch(e){q('#browseList').innerHTML='<div class="empty-msg" style="padding:10px">'+esc(e.message)+'</div>'}}
  function renderBrowse(data){var entries=data.entries||[];q('#browseList').innerHTML=entries.map(function(e){return '<div class="request-row browse-row" data-name="'+esc(e.name)+'" data-dir="'+(e.is_dir?'1':'0')+'" style="grid-template-columns:16px 1fr auto;cursor:pointer;padding:0 10px"><span>'+(e.is_dir?'📁':'💽')+'</span><span class="request-path">'+esc(e.name)+'</span><span class="request-ms">'+(e.is_dir?'':fmtBytes(e.size_bytes))+'</span></div>'}).join('')||'<div class="empty-msg" style="padding:10px">Empty directory.</div>';qa('.browse-row').forEach(function(row){row.onclick=function(){var name=row.dataset.name;if(row.dataset.dir==='1'){browsePath=browsePath?browsePath+'/'+name:name;loadBrowse();return}var rel=browsePath?browsePath+'/'+name:name,vm=q('#uploadVmSelect').value;if(!vm){toast('Pick a VM first','Choose which VM this file belongs to above.');return}assignExistingVMDK(rel,vm)}})}
  async function assignExistingVMDK(fileName,vmName){try{vmdks=await api('/api/vmdks/assign',{method:'POST',body:JSON.stringify({root:browseRoot,file_name:fileName,vm_name:vmName})});renderVmdks();q('#uploadModal').classList.remove('open');toast('VMDK assigned',fileName+' → '+vmName)}catch(e){toast('Assign failed',e.message)}}
  function trend(id,current,previous,suffix,invert){var el=q(id),delta=previous==null?0:Number(current)-Number(previous),good=invert?delta<=0:delta>=0;el.className='metric-trend'+(good?'':' bad');var arrow=delta<0?'↓':(delta>0?'↑':'→');el.textContent=arrow+' '+(Math.abs(delta).toFixed(delta%1?1:0))+(suffix||'')}
  function renderTelemetry(){if(!telemetry)return;q('#kRequests').textContent=Number(telemetry.requests||0).toLocaleString();q('#kError').textContent=Number(telemetry.error_rate_pct||0).toFixed(2)+'%';q('#kSessions').textContent=telemetry.active_sessions||0;q('#kBytes').textContent=fmtBytes(telemetry.bytes_transferred);q('#kExports').textContent=telemetry.exports||0;q('#kLatency').textContent=Math.round(telemetry.average_response_ms||0)+' ms';q('#alertBadge').textContent=Math.min(99,telemetry.errors||0);if(lastTelemetry){trend('#tRequests',telemetry.requests,lastTelemetry.requests,'');trend('#tError',telemetry.error_rate_pct,lastTelemetry.error_rate_pct,'%',true);trend('#tSessions',telemetry.active_sessions,lastTelemetry.active_sessions,'');trend('#tBytes',telemetry.bytes_transferred,lastTelemetry.bytes_transferred,' B');trend('#tExports',telemetry.exports,lastTelemetry.exports,'');trend('#tLatency',telemetry.average_response_ms,lastTelemetry.average_response_ms,' ms',true)}renderActivity();renderRequests();renderHealth();lastTelemetry=JSON.parse(JSON.stringify(telemetry))}
  function renderActivity(){var items=telemetry.activity||[],total=telemetry.requests||0;q('#donutTotal').textContent=Number(total).toLocaleString();if(!items.length){q('#activityLegend').innerHTML='<div class="activity-empty">No traffic yet.</div>';q('#donut').style.setProperty('--parts','conic-gradient(#e8e8ed 0 100%)');return}var acc=0,parts=[],html='';items.forEach(function(x,i){var p=Number(x.percent||0),start=acc,end=acc+p,clr=colors[i%colors.length];parts.push(clr+' '+start.toFixed(2)+'% '+end.toFixed(2)+'%');acc=end;html+='<div class="legend-row"><span class="legend-dot" style="background:'+clr+'"></span><span>'+esc(x.name)+'</span><span>'+Number(x.count).toLocaleString()+' ('+Math.round(p)+'%)</span></div>'});if(acc<100)parts.push('#e8e8ed '+acc.toFixed(2)+'% 100%');q('#donut').style.setProperty('--parts','conic-gradient('+parts.join(',')+')');q('#activityLegend').innerHTML=html}
  var requestRowLimit=12;
  function renderRequests(){var rows=(telemetry.recent||[]).slice(0,requestRowLimit);if(!rows.length){q('#requestList').innerHTML='<div class="request-empty">Waiting for client traffic…</div>';return}q('#requestList').innerHTML=rows.map(function(r){var m=r.method==='GET'?'get':'',st=r.status>=400?'err':'';return '<div class="request-row"><span class="method '+m+'">'+esc(r.method)+'</span><span class="request-path" title="'+esc(r.path)+'">'+esc(r.path)+'</span><span class="request-status '+st+'">'+r.status+'</span><span class="request-ms">'+r.duration_ms+'ms</span></div>'}).join('')}
  function renderHealth(){var err=Number(telemetry.error_rate_pct||0),pct=Math.max(5,Math.round(100-Math.min(70,err*8)));q('#healthPct').textContent=pct;q('#healthRing').style.setProperty('--health',pct);q('#healthLabel').textContent=pct>90?'Healthy':(pct>70?'Degraded':'Faulted')}
  function renderState(){if(!state)return;q('#fLatency').value=state.latency_ms||0;q('#fStatus').value=state.fail_status||503;q('#fFail').value=state.fail_next||0;q('#fNfcFail').value=state.nfc_fail_next||0;q('#fDrop').value=state.nfc_drop_next||0;q('#fDropAfter').value=((state.nfc_drop_after_bytes||0)/1048576);q('#fBandwidth').value=((state.bandwidth_bytes_per_sec||0)/1048576);q('#faultLatencySummary').textContent=(state.latency_ms||0)+' ms';q('#faultBandwidthSummary').textContent=fmtBps(state.bandwidth_bytes_per_sec);q('#faultStatusSummary').textContent=state.fail_status||503}
  async function refresh(){try{var out=await Promise.all([api('/health'),api('/api/telemetry')]);q('#sideUptime').textContent=fmtUptime(out[0].uptime_seconds);telemetry=out[1];renderTelemetry()}catch(e){q('#healthLabel').textContent='Unavailable'}if(token){try{state=await api('/state');renderState()}catch(e){lock()}}}
  async function boot(){try{var out=await Promise.all([api('/api/bootstrap'),api('/api/inventory'),api('/api/vmdks')]);bootstrap=out[0];inventory=out[1];vmdks=out[2];renderBootstrap();renderVMs();renderVmdks()}catch(e){toast('Bootstrap failed',e.message)}renderState();await refresh();setInterval(refresh,2000);setInterval(function(){q('#systemTime').textContent=new Date().toLocaleString()},1000)}
  async function initApp(){var pg=(location.hash||'#home').slice(1);if(pages.indexOf(pg)<0)pg='home';if(!token){showLoginPage();return}showApp();showPage(pg);try{state=await api('/state');q('#unlockBtn').textContent='Lock';q('#lockNote').textContent='✓ Administrative controls unlocked for this tab.';await boot()}catch(e){lock()}}
  qa('.scenario').forEach(function(b){b.onclick=async function(){if(!token){showLoginPage();return}try{state=await api('/scenario/'+b.dataset.scenario,{method:'POST'});renderState();qa('.scenario').forEach(function(x){x.classList.toggle('active',x===b)});toast('Scenario applied',b.dataset.scenario)}catch(e){toast('Scenario failed',e.message)}}});
  q('#applyFaults').onclick=async function(){if(!token){showLoginPage();return}var body={latency_ms:+q('#fLatency').value||0,fail_status:+q('#fStatus').value||503,fail_next:+q('#fFail').value||0,nfc_fail_next:+q('#fNfcFail').value||0,nfc_drop_next:+q('#fDrop').value||0,nfc_drop_after_bytes:Math.round((+q('#fDropAfter').value||0)*1048576),bandwidth_bytes_per_sec:Math.round((+q('#fBandwidth').value||0)*1048576)};try{state=await api('/faults',{method:'POST',body:JSON.stringify(body)});renderState();toast('Fault policy active','Traffic will follow the new policy.')}catch(e){toast('Apply failed',e.message)}};
  q('#resetFaults').onclick=async function(){if(!token){showLoginPage();return}try{state=await api('/reset',{method:'POST'});renderState();toast('Simulation reset','Clean traffic path restored.')}catch(e){toast('Reset failed',e.message)}};
  q('#faultTop').onclick=q('#openFaultsBottom').onclick=function(e){if(e)e.preventDefault();openDrawer()};q('#closeDrawer').onclick=q('#overlay').onclick=closeDrawer;
  function toggleLogin(e){if(e)e.preventDefault();if(token){lock();toast('Logged out','Admin token removed from this tab.')}else showLoginPage()}
  q('#unlockBtn').onclick=q('#userMenu').onclick=toggleLogin;q('#authSave').onclick=function(){login(q('#authUser').value,q('#authPass').value)};q('#authPass').onkeydown=function(e){if(e.key==='Enter')login(q('#authUser').value,e.target.value)};q('#refreshBtn').onclick=refresh;
  q('#gearBtn').onclick=function(){openDrawer('settingsDrawer')};q('#closeSettingsDrawer').onclick=closeDrawer;
  q('#bellBtn').onclick=function(){showPage('telemetry')};
  q('#viewAllRequests').onclick=function(){requestRowLimit=requestRowLimit===12?30:12;q('#viewAllRequests').textContent=requestRowLimit===12?'View all':'View less';renderRequests()};
  var topoZoom=1;
  function updateTopoZoom(){q('.topo-canvas').style.transform='scale('+topoZoom+')'}
  function openTopoFullscreen(){q('#topology').classList.add('fullscreen');q('#topoOverlay').classList.add('open');q('#viewFullTopology').textContent='Exit full screen'}
  function closeTopoFullscreen(){q('#topology').classList.remove('fullscreen');q('#topoOverlay').classList.remove('open');q('#viewFullTopology').textContent='Full screen'}
  q('#viewFullTopology').onclick=function(){q('#topology').classList.contains('fullscreen')?closeTopoFullscreen():openTopoFullscreen()};q('#topoOverlay').onclick=closeTopoFullscreen;
  q('#fitTopology').onclick=function(){topoZoom=1;updateTopoZoom()};q('#zoomTopology').onclick=function(){topoZoom=Math.min(1.6,topoZoom+0.15);updateTopoZoom()};
  document.addEventListener('keydown',function(e){if(e.key==='Escape')closeTopoFullscreen()});
  q('#saveCredentials').onclick=async function(){if(!token){showLoginPage();return}var u=q('#settingsUser').value.trim(),p=q('#settingsPass').value;if(!u||!p){toast('Missing fields','Enter both username and password.');return}try{await api('/admin/credentials',{method:'POST',body:JSON.stringify({username:u,password:p})});q('#settingsUser').value='';q('#settingsPass').value='';toast('Login updated','Use the new credentials next time you log in.')}catch(e){toast('Save failed',e.message)}};
  q('#openUploadBtn').onclick=function(){if(!token){showLoginPage();return}var sel=q('#uploadVmSelect');sel.innerHTML='<option value="">Auto (name-match / round-robin)</option>'+(inventory.virtual_machines||[]).map(function(v){return '<option value="'+esc(v.name)+'">'+esc(v.name)+'</option>'}).join('');q('#uploadFile').value='';var roots=vmdks.roots||[],rootSel=q('#browseRootSelect');rootSel.innerHTML=roots.length?roots.map(function(r,i){return '<option value="'+i+'">'+esc(r)+'</option>'}).join(''):'<option value="">No fixture directories</option>';browseRoot=0;browsePath='';q('#uploadModal').classList.add('open');if(roots.length)loadBrowse();else q('#browseList').innerHTML=''};
  q('#browseRootSelect').onchange=function(){browseRoot=+this.value||0;browsePath='';loadBrowse()};q('#browseUpBtn').onclick=function(){if(!browsePath)return;var parts=browsePath.split('/');parts.pop();browsePath=parts.join('/');loadBrowse()};
  q('#uploadCancel').onclick=function(){q('#uploadModal').classList.remove('open')};
  q('#uploadSubmit').onclick=async function(){var f=q('#uploadFile').files[0];if(!f){toast('No file selected','Choose a .vmdk file first.');return}var fd=new FormData();fd.append('file',f);var vm=q('#uploadVmSelect').value;if(vm)fd.append('vm_name',vm);try{vmdks=await api('/api/vmdks/upload',{method:'POST',body:fd});renderVmdks();q('#uploadModal').classList.remove('open');toast('VMDK uploaded',f.name+(vm?' → '+vm:''))}catch(e){toast('Upload failed',e.message)}};
  q('#vmSearch').oninput=function(){vmPage=1;renderVMs()};q('#powerFilter').onchange=function(){vmPage=1;renderVMs()};
  q('#globalSearch').oninput=function(){q('#vmSearch').value=this.value;vmPage=1;renderVMs();showPage('inventory')};
  q('#copyVmBtn').onclick=function(){var v=filteredVMs()[0];if(v){navigator.clipboard&&navigator.clipboard.writeText(v.name);toast('VM selected',v.name)}else toast('No VM selected','Adjust the filters first.')};
  qa('[data-page]').forEach(function(el){el.onclick=function(e){var pg=el.dataset.page;if(!pg)return;if(el.tagName==='A')e.preventDefault();showPage(pg)}});
  window.addEventListener('hashchange',function(){var pg=(location.hash||'#home').slice(1);if(pages.indexOf(pg)>=0&&q('#appRoot').classList.contains('show'))showPage(pg)});
  document.addEventListener('keydown',function(e){if((e.metaKey||e.ctrlKey)&&e.key.toLowerCase()==='k'){e.preventDefault();showPage('inventory');q('#vmSearch').focus()}});
  q('#systemTime').textContent=new Date().toLocaleString();initApp();
})();
</script>
</body>
</html>`
