<template>
  <Layout>
    <div class="files-page">
      <!-- ===== 顶部栏（桌面端） ===== -->
      <div class="page-header">
        <div class="breadcrumb">
          <button class="crumb-home" @click="navigate('/')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
            {{ t.home }}
          </button>
          <template v-for="(seg, i) in pathSegments" :key="i">
            <span class="crumb-sep">/</span>
            <button class="crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
          </template>
        </div>

        <!-- 普通模式按钮 -->
        <div v-if="!selectMode" class="header-actions">
          <button class="btn-select" @click="enterSelectMode">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>
            {{ t.selectMode }}
          </button>
          <button class="btn-action btn-fetch" @click="showFetch = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/><path d="M3 3h4m0 0V7m0-4L3 7"/><line x1="3" y1="3" x2="7" y2="7"/></svg>
            {{ t.fetchFile }}
          </button>
          <button class="btn-action btn-search" @click="openSearch">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            {{ t.search }}
          </button>
          <button class="btn-action" @click="showFolderUpload = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><polyline points="12 11 12 17"/><polyline points="9 14 12 11 15 14"/></svg>
            {{ t.uploadFolder }}
          </button>
          <input ref="folderInput" type="file" webkitdirectory multiple @change="confirmFolderUpload" style="display:none" />
          <button class="btn-action" @click="showUpload = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            {{ t.upload }}
          </button>
          <button class="btn-action" @click="showMkdir = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
            {{ t.newFolder }}
          </button>
          <button class="btn-action" @click="showCreate = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
            {{ t.newFile }}
          </button>
          <div class="header-divider"></div>
          <button class="btn-settings" @click="showSettings = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
          </button>
          <button class="btn-settings" @click="doLogout" :title="t.logout">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
          </button>
        </div>

        <!-- 选择模式按钮 -->
        <div v-else class="header-actions select-actions">
          <span class="selected-info">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M20 6L9 17l-5-5"/></svg>
            {{ selectedCountText }}
          </span>
          <button class="btn-select-all" @click="toggleSelectAll">{{ allSelected ? t.cancel : t.selectAll }}</button>
          <div class="header-divider"></div>
          <button class="btn-batch btn-batch-danger" :disabled="selected.length === 0" @click="showBatchDelete = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/></svg>
            <span>{{ t.batchDelete }}</span>
          </button>
          <button class="btn-batch" :disabled="selected.length === 0" @click="doBatchDownload">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span>{{ t.batchDownload }}</span>
          </button>
          <button class="btn-batch" :disabled="selected.length === 0" @click="openDirModal('move')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="5 9 2 12 5 15"/><polyline points="9 5 12 2 15 5"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="12" y1="2" x2="12" y2="22"/></svg>
            <span>{{ t.batchMove }}</span>
          </button>
          <button class="btn-batch" :disabled="selected.length === 0" @click="openDirModal('copy')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
            <span>{{ t.batchCopy }}</span>
          </button>
          <button class="btn-batch" :disabled="selected.length === 0" @click="showCompress = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 8l-4-4H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V8z"/><polyline points="17 4 17 8 21 8"/><line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
            <span>{{ t.compress }}</span>
          </button>
          <button class="btn-cancel-select" @click="exitSelectMode">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            {{ t.cancelSelect }}
          </button>
        </div>
      </div>

      <!-- ===== 移动端顶部栏 ===== -->
      <div class="mobile-header">
        <!-- 左：三横线 -->
        <button class="mob-icon-btn" @click="showMobileNav=true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <!-- 中：面包屑 -->
        <div class="mob-breadcrumb">
          <button class="mob-crumb-home" @click="navigate('/')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>
          </button>
          <template v-for="(seg, i) in pathSegments" :key="i">
            <span class="mob-crumb-sep">›</span>
            <button class="mob-crumb-item" @click="navigateToIndex(i)">{{ seg }}</button>
          </template>
        </div>
        <!-- 右：三个点 -->
        <button class="mob-icon-btn" :class="{'select-active': selectMode}" @click="showMobileActions=true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="5" r="1.2" fill="currentColor"/><circle cx="12" cy="12" r="1.2" fill="currentColor"/><circle cx="12" cy="19" r="1.2" fill="currentColor"/></svg>
        </button>
      </div>

      <!-- 移动端左侧导航抽屉 -->
      <teleport to="body">
        <div v-if="showMobileNav" class="mob-drawer-mask" @click="showMobileNav=false">
          <div class="mob-drawer-left" @click.stop>
            <div class="mob-drawer-header">
              <div class="mob-drawer-brand">
                <svg viewBox="0 0 100 100" style="width:28px;height:28px"><defs><linearGradient id="mdlg" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" style="stop-color:#2563EB"/><stop offset="100%" style="stop-color:#38BDF8"/></linearGradient></defs><ellipse cx="50" cy="62" rx="32" ry="18" fill="url(#mdlg)"/><circle cx="36" cy="56" r="16" fill="url(#mdlg)"/><circle cx="58" cy="50" r="20" fill="url(#mdlg)"/><polygon points="50,24 41,42 46,42 46,60 54,60 54,42 59,42" fill="white" opacity="0.95"/></svg>
                <span>CloudOne</span>
              </div>
            </div>
            <nav class="mob-drawer-nav">
              <router-link to="/files" class="mob-nav-item" :class="{active:$route.path.startsWith('/files')}" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"/></svg>
                {{ t.myFiles }}
              </router-link>
              <router-link to="/shared" class="mob-nav-item" :class="{active:$route.path==='/shared'}" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                {{ t.shared }}
              </router-link>
              <router-link to="/public-files" class="mob-nav-item" :class="{active:$route.path==='/public-files'}" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ t.public }}
              </router-link>
              <router-link to="/webdav" class="mob-nav-item" :class="{active:$route.path==='/webdav'}" @click="showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12H3l9-9 9 9h-2"/><path d="M5 12v7a2 2 0 002 2h10a2 2 0 002-2v-7"/><path d="M9 21v-6a2 2 0 012-2h2a2 2 0 012 2v6"/></svg>
                {{ t.webdav }}
              </router-link>
              <div class="mob-drawer-divider"></div>
              <button class="mob-nav-item mob-nav-settings" @click="showSettings=true;showMobileNav=false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
                {{ t.settings }}
              </button>
              <button class="mob-nav-item mob-nav-logout" @click="doLogout">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                {{ t.logout }}
              </button>
            </nav>
          </div>
        </div>

        <!-- 移动端右侧菜单：普通模式 / 选择模式动态切换 -->
        <div v-if="showMobileActions" class="mob-actions-mask" @click="showMobileActions=false">
          <!-- 普通模式菜单 -->
          <div v-if="!selectMode" class="mob-actions-menu" @click.stop>
            <button class="mob-act-item" @click="enterSelectMode();showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>
              {{ t.selectMode }}
            </button>
            <button class="mob-act-item fetch" @click="showFetch=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              {{ t.fetchFile }}
            </button>
            <button class="mob-act-item search" @click="openSearch();showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              {{ t.search }}
            </button>
            <div class="mob-act-divider"></div>
            <button class="mob-act-item" @click="showFolderUpload=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><polyline points="12 11 12 17"/><polyline points="9 14 12 11 15 14"/></svg>
              {{ t.uploadFolder }}
            </button>
            <button class="mob-act-item" @click="showUpload=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
              {{ t.upload }}
            </button>
            <button class="mob-act-item" @click="showMkdir=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
              {{ t.newFolder }}
            </button>
            <button class="mob-act-item" @click="showCreate=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
              {{ t.newFile }}
            </button>
          </div>
          <!-- 选择模式菜单 -->
          <div v-else class="mob-actions-menu" @click.stop>
            <div class="mob-act-select-info">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;flex-shrink:0"><path d="M20 6L9 17l-5-5"/></svg>
              {{ selectedCountText }}
            </div>
            <div class="mob-act-divider"></div>
            <button class="mob-act-item" @click="toggleSelectAll();showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><polyline points="9 11 12 14 22 4"/></svg>
              {{ allSelected ? t.cancel : t.selectAll }}
            </button>
            <button class="mob-act-item danger" :disabled="selected.length===0" @click="showBatchDelete=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/></svg>
              {{ t.batchDelete }}
            </button>
            <button class="mob-act-item" :disabled="selected.length===0" @click="doBatchDownload();showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              {{ t.batchDownload }}
            </button>
            <button class="mob-act-item" :disabled="selected.length===0" @click="openDirModal('move');showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="5 9 2 12 5 15"/><polyline points="9 5 12 2 15 5"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="12" y1="2" x2="12" y2="22"/></svg>
              {{ t.batchMove }}
            </button>
            <button class="mob-act-item" :disabled="selected.length===0" @click="openDirModal('copy');showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
              {{ t.batchCopy }}
            </button>
            <button class="mob-act-item" :disabled="selected.length===0" @click="showCompress=true;showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 8l-4-4H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V8z"/><polyline points="17 4 17 8 21 8"/></svg>
              {{ t.compress }}
            </button>
            <div class="mob-act-divider"></div>
            <button class="mob-act-item cancel" @click="exitSelectMode();showMobileActions=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              {{ t.cancelSelect }}
            </button>
          </div>
        </div>
      </teleport>

      <!-- ===== 文件区域 ===== -->
      <div class="drop-zone" :class="{ dragging }" @dragover.prevent="dragging=true" @dragleave="dragging=false" @drop.prevent="handleDrop" @contextmenu.self.prevent>
        <div v-if="loading" class="loading-state"><div class="spinner-lg"></div></div>
        <div v-else-if="!files.length" class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
          <p>{{ t.noFiles }}</p><span>{{ t.uploadDrop }}</span>
        </div>
        <div v-else class="file-table">
          <div class="file-header" :class="{ 'has-check': selectMode }">
            <div v-if="selectMode" class="col-check" @click="toggleSelectAll">
              <div class="checkmark" :class="{ checked: allSelected, indeterminate: selected.length > 0 && !allSelected }">
                <svg v-if="allSelected" viewBox="0 0 12 12" fill="none"><polyline points="2,6 5,9 10,3" stroke="white" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                <svg v-else-if="selected.length > 0" viewBox="0 0 12 12" fill="none"><line x1="2" y1="6" x2="10" y2="6" stroke="white" stroke-width="2" stroke-linecap="round"/></svg>
              </div>
            </div>
            <span class="col-name">{{ t.name }}</span>
            <span class="col-size">{{ t.size }}</span>
            <span class="col-date">{{ t.modified }}</span>
            <span class="col-perm">{{ t.permissions }}</span>
          </div>
          <transition-group name="slide-up" tag="div">
            <div
              v-for="file in files"
              :key="file.path"
              class="file-row"
              :class="{ selected: selected.includes(file.path), 'select-mode': selectMode }"
              @click="handleRowClick(file)"
              @contextmenu.prevent="openCtxMenu($event, file)"
              @touchstart="onTouchStart($event, file)"
              @touchend="onTouchEnd"
              @touchmove="onTouchMove"
            >
              <div v-if="selectMode" class="col-check" @click.stop="toggleSelect(file)">
                <div class="checkmark" :class="{ checked: selected.includes(file.path) }">
                  <svg v-if="selected.includes(file.path)" viewBox="0 0 12 12" fill="none"><polyline points="2,6 5,9 10,3" stroke="white" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>
                </div>
              </div>
              <div class="col-name">
                <div class="file-icon" :class="file.is_dir ? 'folder-icon' : getExt(file.name)">
                  <svg v-if="file.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                </div>
                <span class="file-name">{{ file.name }}</span>
                <span v-if="file.is_public" class="badge-public">{{ t.isPublic }}</span>
              </div>
              <span class="col-size">{{ file.is_dir ? '—' : formatSize(file.size) }}</span>
              <span class="col-date">{{ formatDate(file.mod_time) }}</span>
              <span class="col-perm"><span class="perm-badge">{{ formatMode(file.mode) }}</span></span>
            </div>
          </transition-group>
        </div>
      </div>
    </div>

    <!-- ===== 右键菜单 ===== -->
    <teleport to="body">
      <div v-if="ctxMenu.show" class="ctx-overlay" @mousedown.self="closeCtxMenu" @contextmenu.prevent>
        <div class="ctx-menu" :style="{ top: ctxMenu.y + 'px', left: ctxMenu.x + 'px' }">
          <div class="ctx-item" @click="ctxAction('edit')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
            <span>{{ ctxMenu.file?.is_dir ? (lang==='zh'?'打开':'Open') : t.editFile }}</span>
          </div>
          <div class="ctx-divider"></div>
          <div v-if="!ctxMenu.file?.is_dir" class="ctx-item" @click="ctxAction('download')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span>{{ t.download }}</span>
          </div>
          <div class="ctx-item" @click="ctxAction('share')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
            <span>{{ t.share }}</span>
          </div>
          <div class="ctx-item" @click="ctxAction('togglePublic')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
            <span>{{ ctxMenu.file?.is_public ? t.setPrivate : t.setPublic }}</span>
          </div>
          <div class="ctx-divider"></div>
          <div class="ctx-item" @click="ctxAction('move')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="5 9 2 12 5 15"/><polyline points="9 5 12 2 15 5"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="12" y1="2" x2="12" y2="22"/></svg>
            <span>{{ t.move }}</span>
          </div>
          <div class="ctx-item" @click="ctxAction('copy')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
            <span>{{ t.copy }}</span>
          </div>
          <div class="ctx-item" @click="ctxAction('compress')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 8l-4-4H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V8z"/><polyline points="17 4 17 8 21 8"/><line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
            <span>{{ t.compress }}</span>
          </div>
          <div v-if="isArchive(ctxMenu.file?.name)" class="ctx-item" @click="ctxAction('extract')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 8l-4-4H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V8z"/><polyline points="17 4 17 8 21 8"/><polyline points="12 18 12 12"/><polyline points="9 15 12 18 15 15"/></svg>
            <span>{{ lang==='zh'?'解压到当前目录':'Extract Here' }}</span>
          </div>
          <div class="ctx-divider"></div>
          <div class="ctx-item" @click="ctxAction('chmod')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
            <span>{{ t.permissions }}</span>
          </div>
          <div class="ctx-item" @click="ctxAction('info')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <span>{{ lang==='zh' ? '信息' : 'Info' }}</span>
          </div>
          <div class="ctx-divider"></div>
          <div class="ctx-item ctx-item-danger" @click="ctxAction('delete')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/></svg>
            <span>{{ t.delete }}</span>
          </div>
        </div>
      </div>

      <!-- ===== 所有弹窗 ===== -->

      <!-- 上传文件：选好后手动确认 -->
      <div v-if="showUpload" class="modal-bg" @click.self="showUpload=false">
        <div class="modal">
          <h3>{{ t.upload }}</h3>
          <div class="upload-area" @click="$refs.fileInput.click()" @dragover.prevent @drop.prevent="dropUploadStage">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            <p>{{ t.uploadDrop }}</p>
          </div>
          <input ref="fileInput" type="file" multiple @change="stageFiles" style="display:none" />
          <div v-if="stagedFiles.length" class="upload-list">
            <div v-for="(f,i) in stagedFiles" :key="i" class="upload-item">
              <span>{{ f.name }}</span>
              <span v-if="uploadDone" class="done">✓</span>
              <span v-else class="pending">待上传</span>
            </div>
          </div>
          <div v-if="uploadProgress.length && !uploadDone" class="compress-progress">
            <div class="spinner-sm"></div><span>{{ lang==='zh'?'上传中...':'Uploading...' }}</span>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showUpload=false;stagedFiles=[];uploadDone=false">{{ t.cancel }}</button>
            <button v-if="stagedFiles.length" class="btn-primary-sm" :disabled="uploadDone" @click="doStagedUpload">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
              {{ lang==='zh'?`上传 (${stagedFiles.length} 个)`:`Upload (${stagedFiles.length})` }}
            </button>
          </div>
        </div>
      </div>

      <!-- 上传文件夹：同样需要手动确认 -->
      <div v-if="showFolderUpload" class="modal-bg" @click.self="showFolderUpload=false">
        <div class="modal">
          <h3>{{ t.uploadFolder }}</h3>
          <div class="upload-area" @click="$refs.folderInput.click()" @dragover.prevent @drop.prevent="dropFolderStage">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/><polyline points="12 11 12 17"/><polyline points="9 14 12 11 15 14"/></svg>
            <p>{{ lang==='zh'?'点击选择文件夹，或将文件夹拖拽到此处':'Click to select folder, or drag & drop a folder here' }}</p>
          </div>
          <div v-if="stagedFolderFiles.length" class="upload-list">
            <div class="upload-item">
              <span>📁 {{ stagedFolderName }} ({{ stagedFolderFiles.length }} {{ lang==='zh'?'个文件':'files' }})</span>
              <span v-if="folderUploadDone" class="done">✓</span>
            </div>
          </div>
          <div v-if="folderUploading" class="compress-progress">
            <div class="spinner-sm"></div><span>{{ lang==='zh'?'上传中...':'Uploading...' }}</span>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showFolderUpload=false;stagedFolderFiles=[];folderUploadDone=false">{{ t.cancel }}</button>
            <button v-if="stagedFolderFiles.length" class="btn-primary-sm" :disabled="folderUploading||folderUploadDone" @click="doFolderUpload">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
              {{ lang==='zh'?`上传文件夹`:`Upload Folder` }}
            </button>
          </div>
        </div>
      </div>

      <!-- 新建文件夹 -->
      <div v-if="showMkdir" class="modal-bg" @click.self="showMkdir=false">
        <div class="modal">
          <h3>{{ t.newFolder }}</h3>
          <div class="field"><label>{{ t.folderName }}</label><input v-model="newDirName" type="text" @keyup.enter="doMkdir" autofocus /></div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showMkdir=false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" @click="doMkdir">{{ t.create }}</button>
          </div>
        </div>
      </div>

      <!-- 新建文件 -->
      <div v-if="showCreate" class="modal-bg" @click.self="showCreate=false">
        <div class="modal modal-lg">
          <h3>{{ t.createFileTitle }}</h3>
          <div class="field"><label>{{ t.fileName }}</label><input v-model="newFile.name" type="text" /></div>
          <div class="field"><label>{{ t.fileContent }}</label><textarea v-model="newFile.content" rows="8"></textarea></div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showCreate=false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" @click="doCreateFile">{{ t.create }}</button>
          </div>
        </div>
      </div>

      <!-- 编辑/预览文件 -->
      <div v-if="showEdit" class="modal-bg" @click.self="showEdit=false">
        <div class="modal modal-xl">
          <div class="modal-titlebar">
            <h3>
              <span v-if="fileViewMode==='image'">{{ lang==='zh'?'预览图片':'Preview' }}</span>
              <span v-else-if="fileViewMode==='unsupported'">{{ lang==='zh'?'无法预览':'Cannot Preview' }}</span>
              <span v-else>{{ t.editFileTitle }}</span>:
              <span class="edit-filename">{{ editTarget?.name }}</span>
            </h3>
            <button class="icon-close" @click="showEdit=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div v-if="fileViewMode==='image'" class="preview-img-wrap"><img :src="previewUrl" :alt="editTarget?.name" class="preview-img" /></div>
          <div v-else-if="fileViewMode==='unsupported'" class="unsupported-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            <p>{{ lang==='zh'?'此文件类型不支持在线预览或编辑':'This file type cannot be previewed or edited online' }}</p>
            <span class="unsupported-ext">.{{ editTarget?.name?.split('.').pop()?.toUpperCase() }}</span>
            <button class="btn-download-hint" @click="downloadFile(editTarget);showEdit=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              {{ t.download }}
            </button>
          </div>
          <div v-else-if="fileViewMode==='text'" class="field edit-field-wrap" style="flex:1">
            <p v-if="editError" class="edit-error">{{ editError }}</p>
            <textarea v-else v-model="editContent" class="code-editor" rows="20" spellcheck="false"></textarea>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showEdit=false">{{ t.cancel }}</button>
            <button v-if="fileViewMode==='text'&&!editError" class="btn-primary-sm" @click="doSaveFile">{{ t.saveFile }}</button>
          </div>
        </div>
      </div>

      <!-- 权限 -->
      <div v-if="showChmod" class="modal-bg" @click.self="showChmod=false">
        <div class="modal">
          <h3>{{ t.permTitle }}: <span class="edit-filename">{{ chmodTarget?.name }}</span></h3>
          <div class="perm-grid">
            <div class="perm-row">
              <span class="perm-label">{{ t.permOwner }}</span>
              <label class="perm-check" :class="{active:perm.ownerR}"><input type="checkbox" v-model="perm.ownerR"/><span>{{ t.permRead }}</span></label>
              <label class="perm-check" :class="{active:perm.ownerW}"><input type="checkbox" v-model="perm.ownerW"/><span>{{ t.permWrite }}</span></label>
              <label class="perm-check" :class="{active:perm.ownerX}"><input type="checkbox" v-model="perm.ownerX"/><span>{{ t.permExec }}</span></label>
            </div>
            <div class="perm-row">
              <span class="perm-label">{{ t.permGroup }}</span>
              <label class="perm-check" :class="{active:perm.groupR}"><input type="checkbox" v-model="perm.groupR"/><span>{{ t.permRead }}</span></label>
              <label class="perm-check" :class="{active:perm.groupW}"><input type="checkbox" v-model="perm.groupW"/><span>{{ t.permWrite }}</span></label>
              <label class="perm-check" :class="{active:perm.groupX}"><input type="checkbox" v-model="perm.groupX"/><span>{{ t.permExec }}</span></label>
            </div>
            <div class="perm-row">
              <span class="perm-label">{{ t.permOther }}</span>
              <label class="perm-check" :class="{active:perm.otherR}"><input type="checkbox" v-model="perm.otherR"/><span>{{ t.permRead }}</span></label>
              <label class="perm-check" :class="{active:perm.otherW}"><input type="checkbox" v-model="perm.otherW"/><span>{{ t.permWrite }}</span></label>
              <label class="perm-check" :class="{active:perm.otherX}"><input type="checkbox" v-model="perm.otherX"/><span>{{ t.permExec }}</span></label>
            </div>
          </div>
          <div class="perm-octal"><span>{{ t.permissions }}:</span><code>{{ permOctal }}</code></div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showChmod=false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" @click="doChmod">{{ t.permApply }}</button>
          </div>
        </div>
      </div>

      <!-- 单个删除确认 -->
      <div v-if="deleteTarget" class="modal-bg" @click.self="deleteTarget=null">
        <div class="modal">
          <h3>{{ t.confirmDelete }}</h3>
          <p class="modal-desc">{{ t.deleteWarning }}</p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="deleteTarget=null">{{ t.cancel }}</button>
            <button class="btn-danger" @click="doDelete">{{ t.confirm }}</button>
          </div>
        </div>
      </div>

      <!-- 文件信息弹窗 -->
      <div v-if="infoTarget" class="modal-bg" @click.self="infoTarget=null">
        <div class="modal info-modal">
          <div class="info-modal-header">
            <div class="info-modal-icon" :class="infoTarget.is_dir ? 'folder-icon' : getExt(infoTarget.name)">
              <svg v-if="infoTarget.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <div class="info-modal-title">
              <span class="info-filename">{{ infoTarget.name }}</span>
              <span class="info-type-badge" :class="infoTarget.is_dir ? 'badge-dir' : 'badge-file'">
                {{ infoTarget.is_dir ? (lang==='zh'?'文件夹':'Folder') : (lang==='zh'?'文件':'File') }}
              </span>
            </div>
            <button class="info-close-btn" @click="infoTarget=null">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="info-list">
            <div class="info-row">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                {{ lang==='zh'?'文件名':'Name' }}
              </span>
              <span class="info-value info-mono">{{ infoTarget.name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
                {{ lang==='zh'?'路径':'Path' }}
              </span>
              <span class="info-value info-mono info-path">{{ infoTarget.path }}</span>
            </div>
            <div class="info-row" v-if="!infoTarget.is_dir">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/></svg>
                {{ lang==='zh'?'大小':'Size' }}
              </span>
              <span class="info-value">{{ formatSize(infoTarget.size) }}<span class="info-bytes"> ({{ infoTarget.size?.toLocaleString() }} bytes)</span></span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
                {{ lang==='zh'?'修改时间':'Modified' }}
              </span>
              <span class="info-value">{{ formatDate(infoTarget.mod_time) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
                {{ lang==='zh'?'权限':'Permissions' }}
              </span>
              <span class="info-value">
                <code class="info-perm-code">{{ formatMode(infoTarget.mode) }}</code>
                <span class="info-perm-text">{{ formatModeStr(infoTarget.mode) }}</span>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
                {{ lang==='zh'?'可见性':'Visibility' }}
              </span>
              <span class="info-value">
                <span class="info-vis-badge" :class="infoTarget.is_public ? 'vis-public' : 'vis-private'">
                  {{ infoTarget.is_public ? (lang==='zh'?'公开':'Public') : (lang==='zh'?'私有':'Private') }}
                </span>
              </span>
            </div>
            <div class="info-row" v-if="!infoTarget.is_dir">
              <span class="info-label">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>
                {{ lang==='zh'?'扩展名':'Extension' }}
              </span>
              <span class="info-value info-mono">{{ infoTarget.name.includes('.') ? '.'+infoTarget.name.split('.').pop().toLowerCase() : (lang==='zh'?'无':'None') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 批量删除确认 -->
      <div v-if="showBatchDelete" class="modal-bg" @click.self="showBatchDelete=false">
        <div class="modal">
          <h3>{{ t.confirmBatchDelete }}</h3>
          <p class="modal-desc">{{ batchDeleteWarningText }}</p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showBatchDelete=false">{{ t.cancel }}</button>
            <button class="btn-danger" @click="doBatchDelete">{{ t.confirm }}</button>
          </div>
        </div>
      </div>

      <!-- 分享 -->
      <div v-if="showShareModal" class="modal-bg" @click.self="showShareModal=false;shareResult=null">
        <div class="modal">
          <h3>{{ t.shareTitle }}</h3>
          <!-- 第一步：尚未创建分享 -->
          <template v-if="!shareResult">
            <p class="modal-desc" style="margin-bottom:8px">
              {{ shareTarget?.name }}
            </p>
            <div class="modal-actions">
              <button class="btn-ghost" @click="showShareModal=false">{{ t.cancel }}</button>
              <button class="btn-primary-sm" @click="doCreateShare">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                {{ lang==='zh'?'保存':'Save' }}
              </button>
            </div>
          </template>
          <!-- 第二步：分享已创建，显示链接 -->
          <template v-else>
            <div class="share-link-box"><input readonly :value="shareUrl" class="share-input"/></div>
            <div class="modal-actions">
              <button class="btn-ghost" @click="showShareModal=false;shareResult=null">{{ t.cancel }}</button>
              <button class="btn-primary-sm" @click="copyShareAndClose">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                {{ t.copyLink }}
              </button>
            </div>
          </template>
        </div>
      </div>

      <!-- 公开/私有 二次确认弹窗 -->
      <div v-if="showVisibilityConfirm" class="modal-bg" @click.self="showVisibilityConfirm=false">
        <div class="modal">
          <h3>{{ visibilityTarget?.is_public ? t.confirmSetPrivate : t.confirmSetPublic }}</h3>
          <p class="modal-desc" style="margin-bottom:6px">
            {{ visibilityTarget?.is_public ? t.confirmSetPrivateDesc : t.confirmSetPublicDesc }}
          </p>
          <div class="visibility-path-box">
            <svg viewBox="0 0 24 24" fill="currentColor" v-if="visibilityTarget?.is_dir" style="width:15px;height:15px;color:#F59E0B;flex-shrink:0"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" v-else style="width:15px;height:15px;color:#3B82F6;flex-shrink:0"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            <code class="visibility-path-text">{{ visibilityTarget?.path }}</code>
          </div>
          <p v-if="!visibilityTarget?.is_public" class="visibility-note">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;flex-shrink:0;color:#F59E0B"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            {{ t.confirmSetPublicNote }}
          </p>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showVisibilityConfirm=false">{{ t.cancel }}</button>
            <button
              :class="visibilityTarget?.is_public ? 'btn-danger' : 'btn-primary-sm'"
              @click="doTogglePublic"
            >
              <svg v-if="!visibilityTarget?.is_public" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/></svg>
              {{ visibilityTarget?.is_public ? t.setPrivate : t.setPublic }}
            </button>
          </div>
        </div>
      </div>

            <!-- 目录选择弹窗（移动 & 复制 共用） -->
      <div v-if="showDirModal" class="modal-bg" @click.self="showDirModal=false">
        <div class="modal modal-move">
          <div class="modal-titlebar">
            <h3>{{ dirModalMode==='move' ? t.moveTitle : t.copyTitle }}</h3>
            <button class="icon-close" @click="showDirModal=false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="move-breadcrumb">
            <button class="move-crumb-btn" @click="loadDirTree('/')">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/></svg>/
            </button>
            <template v-for="(seg,i) in moveBreadcrumb" :key="i">
              <span class="crumb-sep-sm">/</span>
              <button class="move-crumb-btn" @click="loadDirTree('/'+moveBreadcrumb.slice(0,i+1).join('/'))">{{ seg }}</button>
            </template>
          </div>
          <div class="move-target-row">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;color:#10B981;flex-shrink:0"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>
            <span>{{ lang==='zh'?(dirModalMode==='move'?'移动到：':'复制到：'):(dirModalMode==='move'?'Move to:':'Copy to:') }} <strong>{{ moveTargetPath }}</strong></span>
          </div>
          <!-- 禁止移动到自身或子目录的提示 -->
          <div v-if="moveTargetConflict" class="move-conflict-warn">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;flex-shrink:0"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            <span>{{ lang==='zh'?'不能移动到所选文件夹本身或其子目录中':'Cannot move into selected folder or its subdirectories' }}</span>
          </div>
          <div class="dir-tree">
            <div v-if="dirTreeLoading" class="dir-tree-loading"><div class="spinner-sm"></div></div>
            <template v-else>
              <div v-for="dir in dirTreeItems" :key="dir.path" class="dir-tree-item" @click="loadDirTree(dir.path)">
                <svg viewBox="0 0 24 24" fill="currentColor" style="width:16px;height:16px;color:#F59E0B;flex-shrink:0"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <span style="flex:1">{{ dir.name }}</span>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;color:#94a3b8"><polyline points="9 18 15 12 9 6"/></svg>
              </div>
              <div v-if="!dirTreeItems.length" class="dir-tree-empty">{{ lang==='zh'?'无子目录':'No subdirectories' }}</div>
            </template>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showDirModal=false">{{ t.cancel }}</button>
            <button class="btn-primary-sm" :disabled="moveTargetConflict" @click="confirmDirAction">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px">
                <polyline v-if="dirModalMode==='move'" points="5 9 2 12 5 15"/><polyline v-if="dirModalMode==='move'" points="9 5 12 2 15 5"/><line v-if="dirModalMode==='move'" x1="2" y1="12" x2="22" y2="12"/><line v-if="dirModalMode==='move'" x1="12" y1="2" x2="12" y2="22"/>
                <rect v-if="dirModalMode==='copy'" x="9" y="9" width="13" height="13" rx="2"/><path v-if="dirModalMode==='copy'" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
              </svg>
              {{ lang==='zh'?(dirModalMode==='move'?'移动到此处':'复制到此处'):(dirModalMode==='move'?'Move Here':'Copy Here') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 压缩弹窗 -->
      <div v-if="showCompress" class="modal-bg" @click.self="showCompress=false">
        <div class="modal">
          <h3>{{ t.compressTitle }}</h3>
          <div class="field">
            <label>{{ t.compressFormat }}</label>
            <div class="format-options">
              <label v-for="fmt in compressFormats" :key="fmt.value" class="format-option" :class="{active: compressForm.format===fmt.value}">
                <input type="radio" :value="fmt.value" v-model="compressForm.format" style="display:none"/>
                <span class="format-icon">{{ fmt.icon }}</span>
                <span class="format-label">{{ fmt.label }}</span>
                <span class="format-ext">{{ fmt.ext }}</span>
              </label>
            </div>
          </div>
          <div class="field">
            <label>{{ t.compressOutput }}</label>
            <input v-model="compressForm.output" type="text" :placeholder="t.compressOutputPlaceholder" />
          </div>
          <div v-if="compressing" class="compress-progress">
            <div class="spinner-sm"></div><span>{{ lang==='zh'?'正在压缩...':'Compressing...' }}</span>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="exitSelectAndClose('compress')">{{ t.cancel }}</button>
            <button class="btn-primary-sm" :disabled="compressing" @click="doCompress">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M21 8l-4-4H5a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2V8z"/><polyline points="17 4 17 8 21 8"/></svg>
              {{ t.confirm }}
            </button>
          </div>
        </div>
      </div>

      <!-- 搜索弹窗 -->
      <div v-if="showSearch" class="modal-bg" @click.self="showSearch=false">
        <div class="modal">
          <h3>{{ t.searchTitle }}</h3>
          <div class="field">
            <label>{{ t.searchName }}</label>
            <input v-model="searchName" type="text" :placeholder="t.searchNamePlaceholder" autofocus @keyup.enter="doSearch" />
          </div>
          <div class="field">
            <label>{{ t.searchDir }}</label>
            <div class="search-dir-input-wrap">
              <span class="search-dir-prefix">/</span>
              <input
                v-model="searchDirInput"
                type="text"
                class="search-dir-input"
                :placeholder="t.searchDirPlaceholder"
                @keyup.enter="doSearch"
              />
            </div>
            <p class="search-dir-hint">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:12px;height:12px;flex-shrink:0"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              {{ t.searchDirHint }}
            </p>
          </div>
          <div class="search-scope-preview">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;flex-shrink:0;color:var(--blue-500)"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
            <span>{{ t.searchScope }}：</span>
            <strong>{{ searchScopeDisplay }}</strong>
            <span style="color:var(--gray-400)"> {{ t.searchScopeDir }}</span>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showSearch=false;searchName='';searchDirInput=''">{{ t.cancel }}</button>
            <button class="btn-primary-sm" :disabled="searching||!searchName.trim()" @click="doSearch">
              <svg v-if="searching" class="spin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              {{ searching ? t.searching : t.search }}
            </button>
          </div>
        </div>
      </div>

      <!-- 搜索结果弹窗 -->
      <div v-if="showSearchResult" class="modal-bg" @click.self="showSearchResult=false">
        <div class="modal modal-search-result">
          <h3>{{ t.searchResults }}: <span class="search-keyword">{{ searchName }}</span></h3>
          <p class="search-result-scope">{{ t.searchScope }}：<strong>{{ searchUsedDir }}</strong></p>
          <div v-if="searchResults.length === 0" class="search-empty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
            <p>{{ t.searchNoResult }}</p>
          </div>
          <div v-else class="search-result-list">
            <div v-for="r in searchResults" :key="r.path" class="search-result-item" :class="{ selected: searchSelected?.path === r.path }" @click="searchSelected = (searchSelected?.path === r.path ? null : r)">
              <div class="sri-icon" :class="r.is_dir ? 'sri-dir' : 'sri-file'">
                <svg v-if="r.is_dir" viewBox="0 0 24 24" fill="currentColor"><path d="M10 4H4a2 2 0 00-2 2v12a2 2 0 002 2h16a2 2 0 002-2V8a2 2 0 00-2-2h-8l-2-2z"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </div>
              <div class="sri-info">
                <p class="sri-name">{{ r.name }}</p>
                <p class="sri-path">{{ r.path }}</p>
              </div>
              <div v-if="searchSelected?.path === r.path" class="sri-check">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
              </div>
            </div>
          </div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showSearchResult=false; searchSelected=null">{{ t.cancel }}</button>
            <button v-if="searchSelected" class="btn-primary-sm" @click="jumpToSearchResult">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><polyline points="9 18 15 12 9 6"/></svg>
              {{ t.searchJump }}
            </button>
          </div>
        </div>
      </div>

      <!-- 远程获取弹窗 -->
      <div v-if="showFetch" class="modal-bg" @click.self="showFetch=false">
        <div class="modal">
          <h3>{{ t.fetchTitle }}</h3>
          <div class="field">
            <label>{{ t.fetchUrl }}</label>
            <input v-model="fetchForm.url" type="url" :placeholder="t.fetchUrlPlaceholder" autofocus />
          </div>
          <div class="field">
            <label>{{ t.fetchFilename }}</label>
            <input v-model="fetchForm.filename" type="text" :placeholder="t.fetchFilenamePlaceholder" />
          </div>
          <div v-if="fetching" class="compress-progress">
            <div class="spinner-sm"></div><span>{{ t.fetching }}</span>
          </div>
          <div v-if="fetchError" class="fetch-error">{{ fetchError }}</div>
          <div class="modal-actions">
            <button class="btn-ghost" @click="showFetch=false;fetchError=''">{{ t.cancel }}</button>
            <button class="btn-primary-sm" :disabled="fetching||!fetchForm.url" @click="doFetch">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              {{ t.confirm }}
            </button>
          </div>
        </div>
      </div>

    </teleport>
    <SettingsModal v-model="showSettings" @storage-changed="onStorageChanged" />
    <!-- Toast 提示 -->
    <teleport to="body">
      <div v-if="toastMsg" class="toast-tip">{{ toastMsg }}</div>
    </teleport>
  </Layout>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import Layout from '../components/Layout.vue'
import SettingsModal from '../components/SettingsModal.vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import { t, currentLang as lang } from '../i18n'

// ── 状态 ──────────────────────────────────────────
const showSettings = ref(false)
const _router = useRouter()
const _auth = useAuthStore()
function doLogout() { _auth.logout(); _router.push('/login') }
function onStorageChanged() { currentPath.value = '/'; load() }
const currentPath = ref('/')
const files = ref([])
const loading = ref(false)
const dragging = ref(false)

const selectMode = ref(false)
const selected = ref([])

// 移动端菜单
const showMobileNav = ref(false)
const webdavEnabled = ref(false)

async function loadWebDAVStatus() {
  try {
    const { data } = await api.get('/webdav/settings')
    webdavEnabled.value = data.webdav_enabled
  } catch {}
}
const showMobileActions = ref(false)

// 弹窗开关
const showUpload = ref(false)
const showFolderUpload = ref(false)
const showMkdir = ref(false)
const showCreate = ref(false)
const showEdit = ref(false)
const showChmod = ref(false)
const showBatchDelete = ref(false)
const showDirModal = ref(false)
const showCompress = ref(false)
const showFetch = ref(false)
const showSearch = ref(false)
const showSearchResult = ref(false)

// 上传文件（暂存后确认）
const stagedFiles = ref([])
const uploadProgress = ref([])
const uploadDone = ref(false)

// 上传文件夹（暂存后确认）
const stagedFolderFiles = ref([])
const stagedFolderName = ref('')
const folderUploading = ref(false)
const folderUploadDone = ref(false)

// 各弹窗数据
const newDirName = ref('')
const newFile = ref({ name:'', content:'' })
const deleteTarget = ref(null)
const infoTarget = ref(null)
const shareResult = ref(null)
const shareTarget = ref(null)
const showShareModal = ref(false)
const copied = ref(false)
const showVisibilityConfirm = ref(false)
const visibilityTarget = ref(null)

const editTarget = ref(null)
const editContent = ref('')
const editError = ref('')
const fileViewMode = ref('text')
const previewUrl = ref('')

const chmodTarget = ref(null)
const perm = ref({ ownerR:false,ownerW:false,ownerX:false,groupR:false,groupW:false,groupX:false,otherR:false,otherW:false,otherX:false })

const dirModalMode = ref('move')
const moveTargetPath = ref('/')
const moveBreadcrumb = ref([])
const dirTreeItems = ref([])
const dirTreeLoading = ref(false)
// 用于单文件右键移动/复制时临时存储
const singleOpFile = ref(null)

const compressFormats = [
  { value:'zip',    icon:'🗜',  label:'ZIP',    ext:'.zip'    },
  { value:'tar.gz', icon:'📦',  label:'TAR.GZ', ext:'.tar.gz' },
  { value:'tar',    icon:'📁',  label:'TAR',    ext:'.tar'    },
]
const compressForm = ref({ format:'zip', output:'' })
const compressing = ref(false)

const fetchForm = ref({ url:'', filename:'' })
const fetching = ref(false)
const fetchError = ref('')

const searchName = ref('')
const searchDirInput = ref('')  // 用户输入的搜索目录（不含前导 /）
const searching = ref(false)
const searchResults = ref([])
const searchSelected = ref(null)
const searchUsedDir = ref('')  // 本次实际搜索的目录，结果页展示用

// 右键菜单
const ctxMenu = ref({ show:false, x:0, y:0, file:null })

// ── 计算属性 ───────────────────────────────────────
// 实时展示本次将搜索的目录
const searchScopeDisplay = computed(() => {
  const raw = searchDirInput.value.trim()
  if (!raw) return currentPath.value
  // 拼成 /xxx/yyy 形式
  const normalized = '/' + raw.replace(/^\/+/, '').replace(/\/+$/, '')
  return normalized
})

const permOctal = computed(() => {
  const p = perm.value
  return `${(p.ownerR?4:0)+(p.ownerW?2:0)+(p.ownerX?1:0)}${(p.groupR?4:0)+(p.groupW?2:0)+(p.groupX?1:0)}${(p.otherR?4:0)+(p.otherW?2:0)+(p.otherX?1:0)}`
})
const pathSegments = computed(() => currentPath.value.split('/').filter(Boolean))
const shareUrl = computed(() => shareResult.value ? `${location.origin}/s/${shareResult.value.code}` : '')
const rawShareUrl = computed(() => shareResult.value ? `${location.origin}/s/${shareResult.value.code}/raw` : '')
const allSelected = computed(() => files.value.length > 0 && selected.value.length === files.value.length)
const selectedCountText = computed(() => (t.value.selectedCount||'{n} 项已选').replace('{n}', selected.value.length))
const batchDeleteWarningText = computed(() => (t.value.batchDeleteWarning||'将删除 {n} 个项目！').replace('{n}', selected.value.length))

// 移动目标是否与所选路径冲突（移动到自身或子目录）
const moveTargetConflict = computed(() => {
  if (dirModalMode.value !== 'move') return false
  const target = moveTargetPath.value
  // 获取当前操作的路径列表（批量或单个）
  const opPaths = singleOpFile.value ? [singleOpFile.value.path] : selected.value
  return opPaths.some(p => {
    // 目标路径等于该路径本身（移动到自身内部）
    if (target === p) return true
    // 目标路径是该路径的子目录
    if (target.startsWith(p + '/')) return true
    return false
  })
})

// ── 加载 ──────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/files', { params: { path: currentPath.value } })
    files.value = data.files || []
  } catch {}
  loading.value = false
}
function navigate(path) { currentPath.value = path; if (selectMode.value) exitSelectMode(); load() }
function navigateToIndex(i) { navigate('/'+pathSegments.value.slice(0,i+1).join('/')) }

// ── 选择模式 ──────────────────────────────────────
function enterSelectMode() { selectMode.value = true; selected.value = [] }
function exitSelectMode()  { selectMode.value = false; selected.value = [] }
function toggleSelect(file) {
  const idx = selected.value.indexOf(file.path)
  if (idx === -1) selected.value.push(file.path)
  else selected.value.splice(idx, 1)
}
function toggleSelectAll() {
  if (allSelected.value) selected.value = []
  else selected.value = files.value.map(f => f.path)
}
function handleRowClick(file) {
  if (selectMode.value) { toggleSelect(file); return }
  if (file.is_dir) navigate(file.path)
  else editFile(file)
}
function exitSelectAndClose(which) {
  if (which === 'compress') { showCompress.value = false; compressForm.value = { format:'zip', output:'' } }
  exitSelectMode()
}

// ── 右键菜单 ──────────────────────────────────────
function openCtxMenu(e, file) {
  closeCtxMenu()
  // 动态估算菜单高度：基础项 + 可能的「下载」「解压」项
  const baseH = file.is_dir ? 360 : 390
  const extraH = isArchive(file.name) ? 36 : 0
  const menuW = 200, menuH = baseH + extraH
  let x = e.clientX, y = e.clientY
  if (x + menuW > window.innerWidth)  x = window.innerWidth  - menuW - 8
  if (y + menuH > window.innerHeight) y = window.innerHeight - menuH - 8
  if (y < 8) y = 8
  ctxMenu.value = { show:true, x, y, file }
}
function closeCtxMenu() { ctxMenu.value = { show:false, x:0, y:0, file:null } }

// ── 移动端长按 ────────────────────────────────────
let longPressTimer = null
let touchMoved = false
function onTouchStart(e, file) {
  touchMoved = false
  longPressTimer = setTimeout(() => {
    if (!touchMoved) {
      const menuW = 200
      const baseH = file.is_dir ? 360 : 390
      const extraH = isArchive(file.name) ? 36 : 0
      const menuH = baseH + extraH
      const x = window.innerWidth - menuW - 8
      // 固定 y=160
      const y = 160
      ctxMenu.value = { show:true, x, y, file }
    }
  }, 500)
}
function onTouchEnd() {
  clearTimeout(longPressTimer)
}
function onTouchMove() {
  touchMoved = true
  clearTimeout(longPressTimer)
}

async function ctxAction(action) {
  const file = ctxMenu.value.file
  closeCtxMenu()
  if (!file) return
  switch (action) {
    case 'edit':
      if (file.is_dir) navigate(file.path)
      else editFile(file)
      break
    case 'download': downloadFile(file); break
    case 'share': await shareFile(file); break
    case 'togglePublic': togglePublic(file); break
    case 'chmod': await openChmod(file); break
    case 'info': infoTarget.value = file; break
    case 'delete': confirmDelete(file); break
    case 'move':
      singleOpFile.value = file
      selected.value = [file.path]
      await openDirModal('move')
      break
    case 'copy':
      singleOpFile.value = file
      selected.value = [file.path]
      await openDirModal('copy')
      break
    case 'compress':
      singleOpFile.value = file
      selected.value = [file.path]
      compressForm.value = { format:'zip', output:file.name.split('.')[0]||'archive' }
      showCompress.value = true
      break
    case 'extract':
      await doExtract(file)
      break
  }
}

function isArchive(name) {
  if (!name) return false
  const ext = name.toLowerCase()
  return ext.endsWith('.zip') || ext.endsWith('.tar.gz') || ext.endsWith('.tar') || ext.endsWith('.gz') || ext.endsWith('.rar') || ext.endsWith('.7z')
}

async function doExtract(file) {
  try {
    await api.post('/files/extract', { path: file.path, dir: currentPath.value })
    load()
  } catch(e) {
    alert(e.response?.data?.error || (lang.value==='zh'?'解压失败':'Extract failed'))
  }
}

// ── 批量操作 ──────────────────────────────────────
async function doBatchDelete() {
  try { await api.post('/files/batch-delete', { paths: selected.value }) } catch {}
  showBatchDelete.value = false; exitSelectMode(); load()
}

async function doBatchDownload() {
  const token = localStorage.getItem('token')
  const resp = await fetch('/api/files/batch-download', {
    method:'POST', headers:{'Content-Type':'application/json', Authorization:`Bearer ${token}`},
    body: JSON.stringify({ paths: selected.value })
  })
  const blob = await resp.blob()
  const a = document.createElement('a'); a.href = URL.createObjectURL(blob); a.download = 'download.zip'; a.click(); URL.revokeObjectURL(a.href)
}

async function openDirModal(mode) {
  dirModalMode.value = mode
  moveTargetPath.value = '/'; moveBreadcrumb.value = []
  showDirModal.value = true; await loadDirTree('/')
}

async function loadDirTree(path) {
  dirTreeLoading.value = true; moveTargetPath.value = path
  moveBreadcrumb.value = path==='/' ? [] : path.split('/').filter(Boolean)
  try { const {data} = await api.get('/files/dirtree', {params:{path}}); dirTreeItems.value = data.dirs||[] } catch {}
  dirTreeLoading.value = false
}

async function confirmDirAction() {
  if (moveTargetConflict.value) return
  const paths = singleOpFile.value ? [singleOpFile.value.path] : selected.value
  if (dirModalMode.value === 'move') {
    try { await api.post('/files/batch-move', { paths, target: moveTargetPath.value }) } catch {}
  } else {
    try { await api.post('/files/batch-copy', { paths, target: moveTargetPath.value }) } catch {}
  }
  showDirModal.value = false; singleOpFile.value = null; exitSelectMode(); load()
}

async function doCompress() {
  compressing.value = true
  const paths = singleOpFile.value ? [singleOpFile.value.path] : selected.value
  try {
    await api.post('/files/compress', {
      paths,
      format: compressForm.value.format,
      output: compressForm.value.output || 'archive',
      dir: currentPath.value
    })
  } catch {}
  compressing.value = false
  showCompress.value = false
  compressForm.value = { format:'zip', output:'' }
  singleOpFile.value = null
  exitSelectMode(); load()
}

// ── 搜索 ────────────────────────────────────────────
function openSearch() {
  searchName.value = ''
  searchDirInput.value = ''
  searchSelected.value = null
  searchResults.value = []
  showSearch.value = true
}
async function doSearch() {
  if (!searchName.value.trim()) return
  searching.value = true
  // 计算实际搜索目录：留空取当前目录，否则规范化用户输入
  const rawDir = searchDirInput.value.trim()
  const searchDir = rawDir
    ? '/' + rawDir.replace(/^\/+/, '').replace(/\/+$/, '')
    : currentPath.value
  searchUsedDir.value = searchDir
  try {
    const { data } = await api.get('/files/search', { params: { name: searchName.value.trim(), dir: searchDir } })
    searchResults.value = data.results || []
  } catch { searchResults.value = [] }
  searchSelected.value = null
  showSearch.value = false
  showSearchResult.value = true
  searching.value = false
}
function jumpToSearchResult() {
  if (!searchSelected.value) return
  const r = searchSelected.value
  navigate(r.is_dir ? r.path : (r.path.substring(0, r.path.lastIndexOf('/')) || '/'))
  showSearchResult.value = false; searchSelected.value = null
}

// ── 远程获取 ──────────────────────────────────────
async function doFetch() {
  if (!fetchForm.value.url) return
  fetching.value = true; fetchError.value = ''
  try {
    await api.post('/files/fetch-url', { url: fetchForm.value.url, filename: fetchForm.value.filename, dir: currentPath.value })
    showFetch.value = false; fetchForm.value = { url:'', filename:'' }; load()
  } catch (e) {
    fetchError.value = e.response?.data?.error || (lang.value==='zh'?'获取失败，请检查链接':'Fetch failed')
  }
  fetching.value = false
}

// ── 文件操作工具函数 ──────────────────────────────
function formatMode(mode) {
  if (!mode) return '---'
  const m = mode & 0o777
  return `${(m>>6)&7}${(m>>3)&7}${m&7}`
}
function formatModeStr(mode) {
  if (!mode) return ''
  const m = mode & 0o777
  const bits = (n) => [(n&4)?'r':'-',(n&2)?'w':'-',(n&1)?'x':'-'].join('')
  return bits((m>>6)&7) + bits((m>>3)&7) + bits(m&7)
}
function getExt(name) {
  const ext = name.split('.').pop()?.toLowerCase()
  if (['jpg','jpeg','png','gif','webp','svg'].includes(ext)) return 'img'
  if (['mp4','mkv','avi','mov'].includes(ext)) return 'video'
  if (['mp3','flac','wav','aac'].includes(ext)) return 'audio'
  if (['zip','tar','gz','rar','7z'].includes(ext)) return 'archive'
  if (['pdf'].includes(ext)) return 'pdf'
  if (['md','txt','log','ini','conf','yaml','yml','toml','json','sh','bash'].includes(ext)) return 'text'
  if (['js','ts','py','go','java','rs','c','cpp','h'].includes(ext)) return 'code'
  return 'default'
}
function formatSize(bytes) {
  if (!bytes) return '0 B'
  const u=['B','KB','MB','GB'], i=Math.floor(Math.log(bytes)/Math.log(1024))
  return (bytes/Math.pow(1024,i)).toFixed(1)+' '+u[i]
}
function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN',{year:'numeric',month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'})
}
function downloadFile(file) {
  const token = localStorage.getItem('token')
  fetch(`/api/files/download?path=${encodeURIComponent(file.path)}`,{headers:{Authorization:`Bearer ${token}`}})
    .then(r=>r.blob()).then(blob=>{const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download=file.name;a.click();URL.revokeObjectURL(a.href)})
}

const TEXT_EXTS  = new Set(['txt','md','markdown','log','ini','conf','cfg','env','yaml','yml','toml','json','jsonc','json5','html','htm','xml','svg','css','scss','sass','less','js','mjs','cjs','ts','tsx','jsx','vue','py','go','java','rs','c','cpp','cc','h','hpp','sh','bash','zsh','fish','ps1','bat','cmd','rb','php','pl','lua','r','swift','kt','cs','vb','sql','graphql','proto','dockerfile','makefile','csv','tsv','tex','rst','adoc'])
const IMAGE_EXTS = new Set(['jpg','jpeg','png','gif','webp','bmp','ico','tiff','tif','avif','svg'])
function getFileViewMode(filename) {
  if (!filename||!filename.includes('.')) return 'text'
  const ext = filename.split('.').pop().toLowerCase()
  if (IMAGE_EXTS.has(ext)) return 'image'
  if (TEXT_EXTS.has(ext)) return 'text'
  return 'unsupported'
}
async function editFile(file) {
  editTarget.value=file; editContent.value=''; editError.value=''; previewUrl.value=''
  showEdit.value=true; const mode=getFileViewMode(file.name); fileViewMode.value=mode
  if (mode==='image') {
    const token=localStorage.getItem('token')
    try {
      const resp=await fetch(`/api/files/download?path=${encodeURIComponent(file.path)}`,{headers:{Authorization:`Bearer ${token}`}})
      previewUrl.value=URL.createObjectURL(await resp.blob())
    } catch { fileViewMode.value='unsupported' }
    return
  }
  if (mode==='unsupported') return
  try { const {data}=await api.get('/files/content',{params:{path:file.path}}); editContent.value=data.content }
  catch(e) { editError.value=e.response?.data?.error||t.value.binaryFile }
}
async function doSaveFile() {
  try { await api.put('/files/content',{path:editTarget.value.path,content:editContent.value}); showEdit.value=false; load() }
  catch(e) { editError.value=e.response?.data?.error||'Save failed' }
}
async function openChmod(file) {
  chmodTarget.value=file; showChmod.value=true
  try {
    const {data}=await api.get('/files/permission',{params:{path:file.path}}); const m=data.mode
    perm.value={ownerR:!!(m&0o400),ownerW:!!(m&0o200),ownerX:!!(m&0o100),groupR:!!(m&0o040),groupW:!!(m&0o020),groupX:!!(m&0o010),otherR:!!(m&0o004),otherW:!!(m&0o002),otherX:!!(m&0o001)}
  } catch {}
}
async function doChmod() {
  const p=perm.value
  const mode=((p.ownerR?4:0)+(p.ownerW?2:0)+(p.ownerX?1:0))*64+((p.groupR?4:0)+(p.groupW?2:0)+(p.groupX?1:0))*8+((p.otherR?4:0)+(p.otherW?2:0)+(p.otherX?1:0))
  await api.put('/files/permission',{path:chmodTarget.value.path,mode}); showChmod.value=false; load()
}
// 先弹确认框，用户确认后再执行
function togglePublic(file) {
  visibilityTarget.value = file
  showVisibilityConfirm.value = true
}
async function doTogglePublic() {
  const file = visibilityTarget.value
  if (!file) return
  await api.put('/files/visibility', { path: file.path, is_public: !file.is_public })
  showVisibilityConfirm.value = false
  visibilityTarget.value = null
  await load()
}
async function shareFile(file) {
  shareTarget.value = file
  shareResult.value = null
  showShareModal.value = true
}
async function doCreateShare() {
  const { data } = await api.post('/share', { path: shareTarget.value.path, is_dir: shareTarget.value.is_dir })
  shareResult.value = data
}
async function copyShareAndClose() {
  copyText(shareUrl.value)
  showShareModal.value = false
  shareResult.value = null
  showToast(lang.value === 'zh' ? '链接已复制到剪切板' : 'Link copied to clipboard')
}
async function copyShare() { copyText(shareUrl.value); copied.value=true; setTimeout(()=>copied.value=false,2000) }

// 兼容 HTTP（非 HTTPS）环境的复制函数
function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).catch(() => fallbackCopy(text))
  } else {
    fallbackCopy(text)
  }
}
function fallbackCopy(text) {
  const el = document.createElement('textarea')
  el.value = text
  el.style.cssText = 'position:fixed;top:-9999px;left:-9999px'
  document.body.appendChild(el)
  el.select()
  document.execCommand('copy')
  document.body.removeChild(el)
}
function confirmDelete(file) { deleteTarget.value=file }
async function doDelete() { await api.delete('/files',{params:{path:deleteTarget.value.path}}); deleteTarget.value=null; load() }
async function doMkdir() {
  if (!newDirName.value) return
  await api.post('/files/mkdir',{path:currentPath.value.replace(/\/$/,'')+'/'+newDirName.value})
  newDirName.value=''; showMkdir.value=false; load()
}
async function doCreateFile() {
  if (!newFile.value.name) return
  await api.post('/files/create',{path:currentPath.value.replace(/\/$/,'')+'/'+newFile.value.name,content:newFile.value.content})
  newFile.value={name:'',content:''}; showCreate.value=false; load()
}

// ── 上传文件（暂存 → 确认 → 上传）────────────────
function stageFiles(e) {
  stagedFiles.value = Array.from(e.target.files)
  uploadDone.value = false
}
function dropUploadStage(e) {
  stagedFiles.value = Array.from(e.dataTransfer.files)
  uploadDone.value = false
}
async function doStagedUpload() {
  if (!stagedFiles.value.length) return
  uploadProgress.value = stagedFiles.value.map(f => ({ name:f.name, done:false }))
  const form = new FormData()
  stagedFiles.value.forEach(f => form.append('files', f))
  await api.post('/files/upload', form, { params:{ path: currentPath.value } })
  uploadProgress.value.forEach(u => u.done = true)
  uploadDone.value = true
  setTimeout(() => { showUpload.value=false; stagedFiles.value=[]; uploadDone.value=false; load() }, 800)
}

// ── 上传文件夹（暂存 → 确认 → 上传）──────────────
function confirmFolderUpload(e) {
  // 从 input 选择文件夹后，先暂存
  const fs = Array.from(e.target.files)
  if (!fs.length) return
  stagedFolderFiles.value = fs
  stagedFolderName.value = fs[0].webkitRelativePath?.split('/')[0] || '文件夹'
  folderUploadDone.value = false
  showFolderUpload.value = true
}
function dropFolderStage(e) {
  // 拖拽文件夹进弹窗（浏览器限制，实际拖入的是文件列表）
  const fs = Array.from(e.dataTransfer.files)
  if (!fs.length) return
  stagedFolderFiles.value = fs
  stagedFolderName.value = fs[0].webkitRelativePath?.split('/')[0] || (lang.value==='zh'?'拖入的文件':'Dropped files')
  folderUploadDone.value = false
}
async function doFolderUpload() {
  if (!stagedFolderFiles.value.length) return
  folderUploading.value = true
  const form = new FormData()
  stagedFolderFiles.value.forEach(f => {
    form.append('files', f)
    form.append('paths', f.webkitRelativePath || f.name)
  })
  try { await api.post('/files/upload-folder', form, { params:{ path: currentPath.value } }) } catch {}
  folderUploading.value = false
  folderUploadDone.value = true
  setTimeout(() => {
    showFolderUpload.value = false
    stagedFolderFiles.value = []; folderUploadDone.value = false
    if (window.__folderInputRef) window.__folderInputRef.value = ''
    load()
  }, 800)
}

async function handleDrop(e) {
  dragging.value = false
  const fs = Array.from(e.dataTransfer.files); if (!fs.length) return
  const form = new FormData(); fs.forEach(f => form.append('files', f))
  await api.post('/files/upload', form, { params:{ path: currentPath.value } }); load()
}

// 键盘 Esc 关闭右键菜单
function onKeydown(e) { if (e.key === 'Escape') closeCtxMenu() }

// Toast
const toastMsg = ref('')
let toastTimer = null
function showToast(msg) {
  toastMsg.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastMsg.value = '' }, 2500)
}
onMounted(() => { load(); loadWebDAVStatus(); document.addEventListener('keydown', onKeydown) })
onUnmounted(() => { document.removeEventListener('keydown', onKeydown) })
</script>

<style scoped>
.files-page { flex:1; display:flex; flex-direction:column; height:100%; overflow:hidden; }
.page-header { padding:14px 28px; border-bottom:1px solid var(--gray-100); display:flex; align-items:center; justify-content:space-between; background:white; flex-shrink:0; gap:12px; min-height:60px; }
.breadcrumb { display:flex; align-items:center; gap:4px; flex-wrap:wrap; flex-shrink:0; }
.crumb-home,.crumb-item { display:flex; align-items:center; gap:5px; background:none; border:none; color:var(--blue-600); font-size:14px; font-weight:500; font-family:inherit; cursor:pointer; padding:4px 8px; border-radius:6px; transition:var(--transition); }
.crumb-home svg { width:15px; height:15px; }
.crumb-home:hover,.crumb-item:hover { background:var(--blue-50); }
.crumb-sep { color:var(--gray-300); font-size:14px; }
.header-actions { display:flex; align-items:center; gap:6px; flex-wrap:wrap; flex-shrink:0; }
.header-divider { width:1px; height:22px; background:var(--gray-200); margin:0 2px; flex-shrink:0; }
.btn-select { display:flex; align-items:center; gap:5px; padding:7px 13px; border:1.5px solid var(--blue-200); border-radius:var(--radius-sm); background:var(--blue-50); font-size:13px; font-weight:500; font-family:inherit; color:var(--blue-600); cursor:pointer; transition:var(--transition); }
.btn-select svg { width:14px; height:14px; }
.btn-select:hover { background:var(--blue-100); border-color:var(--blue-400); }
.btn-action { display:flex; align-items:center; gap:5px; padding:7px 12px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:white; font-size:13px; font-weight:500; font-family:inherit; color:var(--gray-600); cursor:pointer; transition:var(--transition); }
.btn-action svg { width:15px; height:15px; }
.btn-action:hover { border-color:var(--blue-400); color:var(--blue-600); background:var(--blue-50); }
.btn-fetch { border-color:rgba(16,185,129,0.35); color:#059669; background:rgba(16,185,129,0.06); }
.btn-fetch:hover { border-color:#10B981; color:#047857; background:rgba(16,185,129,0.12); }
.btn-search { border-color:rgba(139,92,246,0.35); color:#7C3AED; background:rgba(139,92,246,0.06); }
.btn-search:hover { border-color:#8B5CF6; color:#6D28D9; background:rgba(139,92,246,0.12); }
.btn-settings { width:34px; height:34px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:white; display:flex; align-items:center; justify-content:center; color:var(--gray-500); cursor:pointer; transition:var(--transition); }
.btn-settings svg { width:16px; height:16px; }
.btn-settings:hover { border-color:var(--blue-400); color:var(--blue-600); background:var(--blue-50); }
.select-actions { animation:fadeSlide 0.2s ease; }
@keyframes fadeSlide { from{opacity:0;transform:translateX(10px)} to{opacity:1;transform:translateX(0)} }
.selected-info { display:flex; align-items:center; gap:5px; font-size:13px; font-weight:600; color:var(--blue-700); background:rgba(37,99,235,0.08); padding:5px 11px; border-radius:20px; white-space:nowrap; }
.btn-select-all { padding:6px 12px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:white; font-size:13px; font-weight:500; font-family:inherit; color:var(--gray-600); cursor:pointer; transition:var(--transition); white-space:nowrap; }
.btn-select-all:hover { background:var(--gray-50); }
.btn-batch { display:flex; align-items:center; gap:5px; padding:7px 11px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:white; font-size:13px; font-weight:500; font-family:inherit; color:var(--gray-700); cursor:pointer; transition:var(--transition); white-space:nowrap; }
.btn-batch svg { width:14px; height:14px; }
.btn-batch:hover:not(:disabled) { border-color:var(--blue-400); color:var(--blue-600); background:var(--blue-50); }
.btn-batch:disabled { opacity:.35; cursor:not-allowed; }
.btn-batch-danger:hover:not(:disabled) { border-color:#fca5a5 !important; color:#dc2626 !important; background:rgba(239,68,68,.06) !important; }
.btn-cancel-select { display:flex; align-items:center; gap:5px; padding:7px 12px; border:none; border-radius:var(--radius-sm); background:var(--gray-100); font-size:13px; font-weight:500; font-family:inherit; color:var(--gray-600); cursor:pointer; transition:var(--transition); white-space:nowrap; }
.btn-cancel-select svg { width:13px; height:13px; }
.btn-cancel-select:hover { background:var(--gray-200); }

/* 文件表格 */
.drop-zone { flex:1; overflow:auto; padding:20px 28px; transition:background .2s; }
.drop-zone.dragging { background:rgba(37,99,235,.04); box-shadow:inset 0 0 0 2px rgba(37,99,235,.3); }
.empty-state { display:flex; flex-direction:column; align-items:center; justify-content:center; height:300px; color:var(--gray-300); gap:12px; }
.empty-state svg { width:64px; height:64px; }
.empty-state p { font-size:16px; font-weight:500; color:var(--gray-400); }
.empty-state span { font-size:13px; color:var(--gray-300); }
.loading-state { display:flex; justify-content:center; padding-top:80px; }
.spinner-lg { width:32px; height:32px; border:3px solid var(--gray-200); border-top-color:var(--blue-500); border-radius:50%; animation:spin .8s linear infinite; }
.spinner-sm { width:18px; height:18px; border:2px solid var(--gray-200); border-top-color:var(--blue-500); border-radius:50%; animation:spin .8s linear infinite; }
@keyframes spin { to{transform:rotate(360deg)} }
.file-table { background:white; border-radius:var(--radius-lg); box-shadow:var(--shadow-sm); overflow:hidden; border:1px solid var(--gray-100); }
.file-header { display:grid; grid-template-columns:1fr 90px 160px 60px; padding:11px 20px; background:var(--gray-50); border-bottom:1px solid var(--gray-100); font-size:12px; font-weight:600; color:var(--gray-400); text-transform:uppercase; letter-spacing:.5px; align-items:center; }
.file-header.has-check { grid-template-columns:44px 1fr 90px 160px 60px; cursor:pointer; }
.file-row { display:grid; grid-template-columns:1fr 90px 160px 60px; padding:11px 20px; border-bottom:1px solid var(--gray-50); cursor:pointer; transition:var(--transition); align-items:center; }
.file-row:last-child { border-bottom:none; }
.file-row:hover { background:var(--blue-50); }
.file-row.select-mode { grid-template-columns:44px 1fr 90px 160px 60px; }
.file-row.selected { background:rgba(37,99,235,.06); }
.file-row.selected:hover { background:rgba(37,99,235,.1); }
.col-check { width:44px; display:flex; align-items:center; justify-content:center; flex-shrink:0; cursor:pointer; }
.checkmark { width:18px; height:18px; border:2px solid var(--gray-300); border-radius:5px; background:white; display:flex; align-items:center; justify-content:center; transition:all .15s; flex-shrink:0; user-select:none; }
.checkmark:hover { border-color:var(--blue-400); }
.checkmark.checked { background:var(--blue-600); border-color:var(--blue-600); }
.checkmark.indeterminate { background:var(--blue-600); border-color:var(--blue-600); }
.col-name { display:flex; align-items:center; gap:10px; min-width:0; }
.file-icon { width:32px; height:32px; border-radius:8px; display:flex; align-items:center; justify-content:center; flex-shrink:0; }
.file-icon.folder-icon { background:rgba(251,191,36,.15); color:#F59E0B; }
.file-icon.img { background:rgba(16,185,129,.12); color:#10B981; }
.file-icon.video { background:rgba(139,92,246,.12); color:#8B5CF6; }
.file-icon.audio { background:rgba(244,63,94,.12); color:#F43F5E; }
.file-icon.archive { background:rgba(245,158,11,.12); color:#F59E0B; }
.file-icon.pdf { background:rgba(239,68,68,.12); color:#EF4444; }
.file-icon.code,.file-icon.text { background:rgba(59,130,246,.12); color:#3B82F6; }
.file-icon.default { background:var(--gray-100); color:var(--gray-400); }
.file-icon svg { width:16px; height:16px; }
.file-name { font-size:14px; font-weight:500; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; color:var(--gray-700); }
.badge-public { padding:2px 8px; background:rgba(16,185,129,.12); color:#10B981; border-radius:20px; font-size:11px; font-weight:600; flex-shrink:0; }
.file-row:hover .file-name { color:var(--blue-600); }
.col-size,.col-date { font-size:13px; color:var(--gray-400); }
.col-perm { font-size:12px; }
.perm-badge { font-family:'JetBrains Mono',monospace; font-size:12px; padding:2px 8px; background:var(--gray-100); color:var(--gray-500); border-radius:6px; }

/* 文件信息弹窗 */
.info-modal { padding: 0 !important; overflow: hidden; }
.info-modal-header {
  display: flex; align-items: center; gap: 14px;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--gray-100);
  background: linear-gradient(135deg, #F8FAFF 0%, #EEF4FF 100%);
  border-radius: 20px 20px 0 0;
  position: relative;
}
.info-modal-icon {
  width: 44px; height: 44px; border-radius: 11px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.info-modal-icon svg { width: 22px; height: 22px; }
.info-modal-title { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 5px; }
.info-filename { font-size: 15px; font-weight: 700; color: var(--gray-800); word-break: break-all; line-height: 1.3; }
.info-type-badge { font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 20px; width: fit-content; }
.badge-dir  { background: rgba(245,158,11,.12); color: #D97706; }
.badge-file { background: rgba(59,130,246,.12);  color: var(--blue-600); }
.info-close-btn {
  width: 28px; height: 28px; border-radius: 50%;
  border: none; background: rgba(239,68,68,.1); color: #EF4444;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: background .15s; flex-shrink: 0;
}
.info-close-btn:hover { background: rgba(239,68,68,.22); }
.info-close-btn svg { width: 13px; height: 13px; }
.info-list { padding: 6px 0 2px; }
.info-row {
  display: flex; align-items: flex-start; gap: 12px;
  padding: 10px 24px; border-bottom: 1px solid var(--gray-50);
  transition: background .1s;
}
.info-row:last-child { border-bottom: none; }
.info-row:hover { background: var(--gray-50); }
.info-label {
  display: flex; align-items: center; gap: 6px;
  min-width: 86px; font-size: 12px; font-weight: 600;
  color: var(--gray-400); padding-top: 1px; flex-shrink: 0;
}
.info-label svg { width: 12px; height: 12px; flex-shrink: 0; }
.info-value {
  font-size: 13px; color: var(--gray-700); font-weight: 500;
  word-break: break-all; display: flex; align-items: center; gap: 7px; flex-wrap: wrap;
}
.info-mono  { font-family: 'JetBrains Mono', monospace; font-size: 12px; }
.info-path  { color: var(--blue-600); }
.info-bytes { font-size: 11px; color: var(--gray-400); font-weight: 400; }
.info-perm-code {
  font-family: 'JetBrains Mono', monospace; font-size: 14px; font-weight: 700;
  color: var(--gray-700); background: var(--gray-100); padding: 2px 10px; border-radius: 6px;
}
.info-perm-text { font-family: 'JetBrains Mono', monospace; font-size: 12px; color: var(--gray-500); }
.info-vis-badge { font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: 20px; }
.vis-public  { background: rgba(16,185,129,.12); color: #059669; }
.vis-private { background: var(--gray-100); color: var(--gray-500); }

/* 右键菜单 */
.ctx-overlay { position:fixed; inset:0; z-index:300; }
.ctx-menu { position:fixed; background:white; border-radius:12px; box-shadow:0 8px 32px rgba(0,0,0,0.18),0 0 0 1px rgba(0,0,0,0.06); padding:5px; min-width:190px; z-index:301; animation:ctxIn .12s cubic-bezier(.4,0,.2,1); }
@keyframes ctxIn { from{opacity:0;transform:scale(.95)} to{opacity:1;transform:scale(1)} }
.ctx-item { display:flex; align-items:center; gap:9px; padding:8px 12px; border-radius:8px; cursor:pointer; font-size:13px; color:var(--gray-700); transition:background .1s; user-select:none; }
.ctx-item svg { width:15px; height:15px; color:var(--gray-400); flex-shrink:0; }
.ctx-item:hover { background:var(--blue-50); color:var(--blue-700); }
.ctx-item:hover svg { color:var(--blue-500); }
.ctx-item-danger:hover { background:rgba(239,68,68,.08); color:#dc2626; }
.ctx-item-danger:hover svg { color:#EF4444; }
.ctx-divider { height:1px; background:var(--gray-100); margin:4px 0; }

/* 弹窗通用 */
.modal-bg { position:fixed; inset:0; background:rgba(15,23,42,.45); backdrop-filter:blur(4px); display:flex; align-items:center; justify-content:center; z-index:100; }
.modal { background:white; border-radius:20px; padding:32px; width:440px; max-width:90vw; box-shadow:var(--shadow-lg); animation:modalIn .2s cubic-bezier(.4,0,.2,1); overflow:hidden; }
.modal-lg { width:520px; }
.modal-xl { width:760px; max-height:90vh; display:flex; flex-direction:column; padding:0; background:#F8FAFC; }
.modal-xl .modal-titlebar { padding:20px 24px 18px; border-bottom:2px solid var(--gray-200); margin-bottom:0; background:#F1F5F9; border-radius:20px 20px 0 0; }
.modal-xl .field { padding:0; margin:0; background:#F8FAFC; flex:1; display:flex; flex-direction:column; }
.modal-xl .modal-actions { padding:14px 24px; border-top:2px solid var(--gray-200); background:#F1F5F9; margin-top:0; border-radius:0 0 20px 20px; }
.modal-move { width:500px; display:flex; flex-direction:column; }
@keyframes modalIn { from{opacity:0;transform:scale(.95) translateY(8px)} to{opacity:1;transform:scale(1) translateY(0)} }
.modal-titlebar { display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
.modal-titlebar h3 { font-size:17px; font-weight:600; color:var(--gray-800); margin:0; }
.icon-close { width:30px; height:30px; border:none; background:var(--gray-100); border-radius:8px; cursor:pointer; display:flex; align-items:center; justify-content:center; color:var(--gray-500); }
.icon-close svg { width:15px; height:15px; }
.icon-close:hover { background:var(--gray-200); }
.edit-filename { font-weight:400; color:var(--gray-500); font-family:'JetBrains Mono',monospace; font-size:14px; }
.edit-error { font-size:14px; color:#EF4444; padding:12px; background:rgba(239,68,68,.08); border-radius:8px; margin-bottom:16px; }
.modal h3 { font-size:18px; font-weight:600; color:var(--gray-800); margin-bottom:20px; }
.modal-desc { font-size:14px; color:var(--gray-500); margin-bottom:24px; }
.modal-actions { display:flex; gap:10px; justify-content:flex-end; margin-top:20px; flex-shrink:0; }
.btn-ghost { padding:9px 18px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:transparent; color:var(--gray-600); font-size:14px; font-weight:500; font-family:inherit; cursor:pointer; transition:var(--transition); }
.btn-ghost:hover { background:var(--gray-50); }
.btn-primary-sm { display:flex; align-items:center; gap:6px; padding:9px 18px; background:var(--primary-gradient); color:white; border:none; border-radius:var(--radius-sm); font-size:14px; font-weight:600; font-family:inherit; cursor:pointer; transition:var(--transition); box-shadow:var(--primary-shadow); }
.btn-primary-sm svg { width:14px; height:14px; }
.btn-primary-sm:hover { transform:translateY(-1px); }
.btn-primary-sm:disabled { opacity:.5; cursor:not-allowed; transform:none; }
.btn-danger { padding:9px 18px; background:#EF4444; color:white; border:none; border-radius:var(--radius-sm); font-size:14px; font-weight:600; font-family:inherit; cursor:pointer; }
.field { margin-bottom:16px; display:flex; flex-direction:column; }
.field label { display:block; font-size:13px; font-weight:500; color:var(--gray-600); margin-bottom:8px; }
.field input,.field textarea:not(.code-editor) { width:100%; padding:10px 14px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); font-size:14px; font-family:inherit; color:var(--gray-800); outline:none; resize:vertical; transition:var(--transition); box-sizing:border-box; }
.field input:focus,.field textarea:not(.code-editor):focus { border-color:var(--blue-500); box-shadow:0 0 0 3px rgba(59,130,246,.1); }
.edit-field-wrap { padding:16px 20px; flex:1; display:flex; flex-direction:column; background:#F1F5F9; overflow:hidden; }
.code-editor { font-family:'JetBrains Mono','Courier New',monospace; font-size:13px; line-height:1.6; flex:1; min-height:300px; background:white !important; color:#1E293B !important; border:1.5px solid var(--gray-200); border-radius:10px; padding:16px; resize:none; outline:none; width:100%; box-sizing:border-box; transition:border-color .2s; box-shadow:0 2px 8px rgba(15,23,42,0.08); }
.code-editor:focus { border-color:var(--blue-500); box-shadow:0 0 0 3px rgba(59,130,246,.15); }
.upload-area { border:2px dashed var(--gray-200); border-radius:var(--radius); padding:36px; text-align:center; cursor:pointer; transition:var(--transition); margin-bottom:14px; }
.upload-area:hover { border-color:var(--blue-400); background:var(--blue-50); }
.upload-area svg { width:40px; height:40px; color:var(--gray-300); margin-bottom:12px; }
.upload-area p { font-size:14px; color:var(--gray-400); }
.upload-list { margin-bottom:14px; max-height:150px; overflow:auto; }
.upload-item { display:flex; justify-content:space-between; padding:5px 0; font-size:13px; color:var(--gray-600); border-bottom:1px solid var(--gray-50); }
.upload-item .done { color:#10B981; }
.upload-item .pending { color:var(--gray-400); }
.preview-img-wrap { display:flex; justify-content:center; align-items:center; padding:20px; flex:1; min-height:200px; background:var(--gray-50); border-radius:12px; }
.preview-img { max-width:100%; max-height:60vh; object-fit:contain; border-radius:8px; box-shadow:var(--shadow-sm); }
.unsupported-wrap { display:flex; flex-direction:column; align-items:center; justify-content:center; padding:48px; gap:12px; flex:1; }
.unsupported-wrap svg { width:56px; height:56px; color:var(--gray-300); }
.unsupported-wrap p { font-size:14px; color:var(--gray-500); font-weight:500; }
.unsupported-ext { padding:4px 14px; background:var(--gray-100); color:var(--gray-500); border-radius:20px; font-family:'JetBrains Mono',monospace; font-size:13px; font-weight:600; }
.btn-download-hint { display:flex; align-items:center; gap:6px; margin-top:8px; padding:10px 20px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); background:white; color:var(--gray-600); font-size:14px; font-weight:500; font-family:inherit; cursor:pointer; transition:var(--transition); }
.btn-download-hint:hover { border-color:var(--blue-400); color:var(--blue-600); background:var(--blue-50); }
.perm-grid { display:flex; flex-direction:column; gap:10px; margin-bottom:14px; }
.perm-row { display:flex; align-items:center; gap:10px; padding:10px 14px; background:var(--gray-50); border-radius:10px; }
.perm-label { font-size:13px; font-weight:600; color:var(--gray-600); width:72px; flex-shrink:0; }
.perm-check { display:flex; align-items:center; gap:6px; cursor:pointer; padding:5px 10px; border-radius:20px; border:1.5px solid var(--gray-200); background:white; transition:var(--transition); user-select:none; }
.perm-check input { width:13px; height:13px; accent-color:var(--blue-600); cursor:pointer; }
.perm-check span { font-size:12px; font-weight:500; color:var(--gray-500); }
.perm-check.active { border-color:var(--blue-400); background:var(--blue-50); }
.perm-check.active span { color:var(--blue-600); }
.perm-octal { display:flex; align-items:center; gap:10px; font-size:13px; color:var(--gray-500); padding:6px 0; }
.perm-octal code { font-family:'JetBrains Mono',monospace; font-size:15px; font-weight:600; color:var(--gray-700); background:var(--gray-100); padding:4px 12px; border-radius:8px; }
.share-link-box { margin-bottom:10px; }
.share-input { width:100%; padding:10px 14px; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); font-size:13px; font-family:'JetBrains Mono',monospace; color:var(--gray-600); background:var(--gray-50); outline:none; box-sizing:border-box; }
.share-raw { font-size:12px; color:var(--gray-400); margin-bottom:16px; }
.share-raw a { color:var(--blue-500); text-decoration:none; }
.move-breadcrumb { display:flex; align-items:center; gap:3px; padding:8px 12px; background:var(--gray-50); border-radius:8px; margin-bottom:10px; flex-wrap:wrap; min-height:36px; }
.move-crumb-btn { display:flex; align-items:center; gap:3px; background:none; border:none; color:var(--blue-600); font-size:13px; font-weight:500; font-family:inherit; cursor:pointer; padding:3px 7px; border-radius:5px; transition:var(--transition); }
.move-crumb-btn:hover { background:var(--blue-100); }
.crumb-sep-sm { color:var(--gray-300); font-size:13px; }
.move-target-row { display:flex; align-items:center; gap:8px; font-size:13px; color:var(--gray-600); margin-bottom:10px; padding:8px 12px; background:rgba(16,185,129,.06); border-radius:8px; border:1px solid rgba(16,185,129,.18); }
.move-conflict-warn { display:flex; align-items:center; gap:8px; font-size:12px; color:#b45309; background:#fffbeb; border:1px solid #fde68a; border-radius:8px; padding:8px 12px; margin-bottom:10px; }
.dir-tree { border:1.5px solid var(--gray-200); border-radius:10px; max-height:240px; overflow-y:auto; background:white; margin-bottom:4px; }
.dir-tree-loading { display:flex; justify-content:center; padding:28px; }
.dir-tree-item { display:flex; align-items:center; gap:10px; padding:11px 16px; cursor:pointer; transition:background .15s; border-bottom:1px solid var(--gray-50); font-size:14px; color:var(--gray-700); }
.dir-tree-item:last-child { border-bottom:none; }
.dir-tree-item:hover { background:var(--blue-50); color:var(--blue-700); }
.dir-tree-empty { display:flex; justify-content:center; padding:28px; font-size:13px; color:var(--gray-400); }
.format-options { display:flex; gap:10px; flex-wrap:wrap; }
.format-option { display:flex; align-items:center; gap:8px; padding:10px 14px; border:2px solid var(--gray-200); border-radius:10px; cursor:pointer; transition:var(--transition); background:white; flex:1; min-width:100px; }
.format-option:hover { border-color:var(--blue-300); background:var(--blue-50); }
.format-option.active { border-color:var(--blue-500); background:rgba(37,99,235,.06); }
.format-icon { font-size:18px; }
.format-label { font-size:13px; font-weight:600; color:var(--gray-700); }
.format-ext { font-size:11px; color:var(--gray-400); font-family:'JetBrains Mono',monospace; margin-left:auto; }
.format-option.active .format-label { color:var(--blue-700); }
.format-option.active .format-ext { color:var(--blue-500); }
.compress-progress { display:flex; align-items:center; gap:10px; padding:10px 14px; background:var(--blue-50); border-radius:8px; font-size:13px; color:var(--blue-700); margin-top:12px; }
.fetch-error { margin-top:10px; padding:10px 14px; background:rgba(239,68,68,.08); border-radius:8px; font-size:13px; color:#dc2626; }
.search-scope-hint { font-size:12px; color:var(--gray-400); margin:6px 0 4px; display:flex; align-items:center; gap:3px; }
.search-scope-hint strong { color:var(--gray-600); font-weight:600; }
/* 搜索目录输入框 */
.search-dir-input-wrap { display:flex; align-items:center; border:1.5px solid var(--gray-200); border-radius:var(--radius-sm); overflow:hidden; transition:var(--transition); }
.search-dir-input-wrap:focus-within { border-color:var(--blue-500); box-shadow:0 0 0 3px rgba(59,130,246,.1); }
.search-dir-prefix { padding:0 10px; font-family:'JetBrains Mono',monospace; font-size:14px; font-weight:600; color:var(--blue-500); background:var(--blue-50); border-right:1.5px solid var(--gray-200); height:40px; display:flex; align-items:center; flex-shrink:0; }
.search-dir-input { flex:1; padding:10px 12px; border:none; outline:none; font-size:14px; font-family:'JetBrains Mono',monospace; color:var(--gray-800); background:white; min-width:0; }
.search-dir-hint { display:flex; align-items:flex-start; gap:5px; font-size:12px; color:var(--gray-400); margin-top:6px; line-height:1.5; }
/* 实时预览区域 */
.search-scope-preview { display:flex; align-items:center; gap:5px; font-size:12px; color:var(--gray-500); background:var(--blue-50); border:1px solid rgba(59,130,246,.15); border-radius:8px; padding:8px 12px; margin-bottom:4px; flex-wrap:wrap; }
.search-scope-preview strong { color:var(--blue-700); font-family:'JetBrains Mono',monospace; word-break:break-all; }
/* 结果页搜索范围提示 */
.search-result-scope { font-size:12px; color:var(--gray-400); margin:-10px 0 10px; }
.search-result-scope strong { font-family:'JetBrains Mono',monospace; color:var(--gray-600); }
.search-keyword { font-weight:700; color:var(--blue-600); }
.spin-icon { animation: spin 1s linear infinite; }
.modal-search-result { max-height:70vh; display:flex; flex-direction:column; }
.search-empty { display:flex; flex-direction:column; align-items:center; justify-content:center; padding:40px 20px; gap:12px; flex:1; }
.search-empty svg { width:48px; height:48px; color:var(--gray-300); }
.search-empty p { font-size:14px; color:var(--gray-400); font-weight:500; }
.search-result-list { flex:1; overflow-y:auto; display:flex; flex-direction:column; gap:6px; padding:2px 0; max-height:380px; }
.search-result-item { display:flex; align-items:center; gap:12px; padding:10px 12px; border:1.5px solid var(--gray-100); border-radius:10px; cursor:pointer; transition:var(--transition); background:white; }
.search-result-item:hover { border-color:var(--blue-300); background:var(--blue-50); }
.search-result-item.selected { border-color:var(--blue-500); background:var(--blue-50); box-shadow:0 0 0 3px var(--blue-100); }
.sri-icon { width:36px; height:36px; border-radius:8px; display:flex; align-items:center; justify-content:center; flex-shrink:0; }
.sri-dir { background:rgba(251,191,36,0.15); color:#F59E0B; }
.sri-file { background:var(--blue-50); color:var(--blue-500); }
.sri-icon svg { width:18px; height:18px; }
.sri-info { flex:1; min-width:0; }
.sri-name { font-size:13px; font-weight:600; color:var(--gray-800); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.sri-path { font-size:11px; color:var(--gray-400); font-family:'JetBrains Mono',monospace; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; margin-top:2px; }
.sri-check { width:22px; height:22px; background:var(--blue-500); border-radius:50%; display:flex; align-items:center; justify-content:center; flex-shrink:0; }
.sri-check svg { width:12px; height:12px; stroke:white; }
.slide-up-enter-active { transition:all .18s ease; }
.slide-up-enter-from { opacity:0; transform:translateY(-6px); }

/* 公开确认弹窗 */
.visibility-path-box { display:flex; align-items:center; gap:8px; padding:10px 14px; background:var(--gray-50); border:1.5px solid var(--gray-200); border-radius:10px; margin-bottom:12px; }
.visibility-path-text { font-family:'JetBrains Mono',monospace; font-size:13px; color:var(--gray-700); word-break:break-all; }
.visibility-note { display:flex; align-items:flex-start; gap:7px; font-size:12px; color:#92400e; background:#fffbeb; border:1px solid #fde68a; border-radius:8px; padding:9px 12px; margin-bottom:4px; line-height:1.5; }

/* ===================== 移动端适配 ===================== */

/* 移动端顶部栏默认隐藏 */
.mobile-header { display:none; }

@media (max-width: 768px) {
  /* 隐藏桌面端头部 */
  .page-header { display:none; }

  /* 移动端顶部栏 */
  .mobile-header {
    display:flex;
    align-items:center;
    padding:10px 14px;
    background:white;
    border-bottom:1px solid var(--gray-100);
    gap:8px;
    flex-shrink:0;
    min-height:52px;
  }
  .mob-icon-btn {
    width:38px; height:38px;
    border:none; background:none;
    color:var(--gray-600);
    display:flex; align-items:center; justify-content:center;
    border-radius:8px;
    cursor:pointer;
    flex-shrink:0;
    transition:background .15s;
  }
  .mob-icon-btn:active { background:var(--gray-100); }
  .mob-icon-btn svg { width:22px; height:22px; }

  /* 选择模式时三个点高亮 */
  .mob-icon-btn.select-active { color:var(--blue-600); background:var(--blue-50); }

  /* 面包屑 */
  .mob-breadcrumb {
    flex:1; display:flex; align-items:center;
    gap:4px; overflow:hidden; min-width:0;
  }
  .mob-crumb-home {
    display:flex; align-items:center;
    border:none; background:none;
    color:var(--blue-600); cursor:pointer;
    padding:4px; border-radius:6px;
    flex-shrink:0;
  }
  .mob-crumb-sep { color:var(--gray-400); font-size:14px; flex-shrink:0; }
  .mob-crumb-item {
    border:none; background:none;
    color:var(--blue-600); font-size:13px; font-weight:500;
    font-family:inherit; cursor:pointer;
    white-space:nowrap; overflow:hidden; text-overflow:ellipsis;
    max-width:120px; padding:4px 4px; border-radius:6px;
  }

  /* 左侧抽屉遮罩 */
  .mob-drawer-mask {
    position:fixed; inset:0;
    background:rgba(15,23,42,.4);
    backdrop-filter:blur(2px);
    z-index:200;
    display:flex;
  }
  .mob-drawer-left {
    width:260px; max-width:80vw;
    background:white;
    height:100%;
    display:flex; flex-direction:column;
    box-shadow:4px 0 24px rgba(15,23,42,.15);
    animation:slideInLeft .2s ease;
  }
  @keyframes slideInLeft {
    from { transform:translateX(-100%); }
    to   { transform:translateX(0); }
  }
  .mob-drawer-header {
    padding:24px 20px 16px;
    border-bottom:1px solid var(--gray-100);
  }
  .mob-drawer-brand {
    display:flex; align-items:center; gap:10px;
  }
  .mob-drawer-brand span {
    font-size:18px; font-weight:700;
    background:linear-gradient(135deg,#2563EB,#38BDF8);
    -webkit-background-clip:text; -webkit-text-fill-color:transparent;
  }
  .mob-drawer-nav {
    flex:1; padding:12px 10px;
    display:flex; flex-direction:column; gap:2px;
    overflow-y:auto;
  }
  .mob-nav-item {
    display:flex; align-items:center; gap:12px;
    padding:13px 14px;
    border-radius:10px;
    text-decoration:none;
    font-size:15px; font-weight:500;
    color:var(--gray-600);
    border:none; background:none;
    font-family:inherit; cursor:pointer;
    transition:background .15s, color .15s;
    text-align:left;
  }
  .mob-nav-item svg { width:20px; height:20px; flex-shrink:0; }
  .mob-nav-item:hover, .mob-nav-item:active { background:var(--blue-50); color:var(--blue-600); }
  .mob-nav-item.active { background:var(--blue-50); color:var(--blue-600); font-weight:600; }
  .mob-nav-logout { color:#EF4444; }
  .mob-nav-logout:hover { background:rgba(239,68,68,.08) !important; color:#EF4444 !important; }
  .mob-drawer-divider { height:1px; background:var(--gray-100); margin:8px 4px; }

  /* 右侧操作菜单 */
  .mob-actions-mask {
    position:fixed; inset:0;
    background:rgba(15,23,42,.3);
    backdrop-filter:blur(2px);
    z-index:200;
    display:flex; justify-content:flex-end; align-items:flex-start;
  }
  .mob-actions-menu {
    margin-top:52px; margin-right:8px;
    background:white;
    border-radius:14px;
    box-shadow:0 8px 32px rgba(15,23,42,.18), 0 2px 8px rgba(15,23,42,.1);
    min-width:180px;
    overflow:hidden;
    animation:fadeInDown .15s ease;
  }
  @keyframes fadeInDown {
    from { opacity:0; transform:translateY(-8px); }
    to   { opacity:1; transform:translateY(0); }
  }
  .mob-act-item {
    display:flex; align-items:center; gap:12px;
    width:100%; padding:13px 18px;
    border:none; background:none;
    font-size:14px; font-weight:500;
    font-family:inherit; cursor:pointer;
    color:var(--gray-700);
    text-align:left;
    transition:background .12s;
  }
  .mob-act-item svg { width:18px; height:18px; flex-shrink:0; color:var(--gray-500); }
  .mob-act-item:hover, .mob-act-item:active { background:var(--gray-50); }
  .mob-act-item:disabled { opacity:.4; pointer-events:none; }
  .mob-act-item.fetch svg { color:#059669; }
  .mob-act-item.search svg { color:#7C3AED; }
  .mob-act-item.danger { color:#EF4444; }
  .mob-act-item.danger svg { color:#EF4444; }
  .mob-act-item.cancel { color:var(--gray-500); }
  .mob-act-divider { height:1px; background:var(--gray-100); margin:4px 0; }
  .mob-act-select-info {
    display:flex; align-items:center; gap:8px;
    padding:12px 18px 8px;
    font-size:13px; font-weight:600; color:var(--blue-700);
  }

  /* 文件表格 */
  .file-header { grid-template-columns:1fr 70px 44px !important; padding:8px 12px; }
  .file-row { grid-template-columns:1fr 70px 44px !important; padding:10px 12px; }
  .file-header > *:nth-child(3),
  .file-row > *:nth-child(3) { display:none; }
  .file-name { font-size:13px; }
  .file-size { font-size:12px; }
  .row-actions .act-btn:not(.act-more) { display:none; }
  .row-actions { justify-content:flex-end; }

  /* 弹窗底部弹出 */
  .modal-bg { align-items:flex-end; }
  .modal {
    width:100% !important; max-width:100% !important;
    border-radius:20px 20px 0 0 !important;
    padding:24px 20px 32px !important;
    max-height:90vh; overflow-y:auto;
  }
  .modal-xl {
    width:100% !important; max-width:100% !important;
    max-height:92vh !important;
    border-radius:20px 20px 0 0 !important;
  }
  .code-editor { min-height:200px; font-size:14px; }
  .modal .field input { font-size:16px; }
  .dir-tree { max-height:200px; }
  .ctx-menu { min-width:180px; }
  .ctx-item { padding:12px 16px; font-size:14px; }
  .info-modal-header { padding: 16px 16px 14px; }
  .info-row { padding: 10px 16px; }
  .info-label { min-width: 72px; }

  /* 搜索目录输入框移动端适配 */
  .search-dir-input { font-size:16px; } /* 防止 iOS 自动缩放 */
  .search-dir-prefix { font-size:16px; height:44px; padding:0 10px; }
  .search-scope-preview { padding:8px 10px; gap:4px; }
  .search-scope-preview strong { font-size:11px; word-break:break-all; }
  .search-dir-hint { font-size:11px; }
  .search-result-scope { margin:0 0 10px; }
}

@media (max-width: 480px) {
  .file-header { grid-template-columns:1fr 44px !important; }
  .file-row { grid-template-columns:1fr 44px !important; }
  .file-header > *:nth-child(2),
  .file-row > *:nth-child(2) { display:none; }
  .file-icon { width:28px; height:28px; border-radius:6px; }
  .file-icon svg { width:13px; height:13px; }
  .mob-crumb-item { max-width:80px; }
  /* 极小屏隐藏预览行的 Scope 文字，只留路径 */
  .search-scope-preview > span:first-of-type { display:none; }
}
</style>
