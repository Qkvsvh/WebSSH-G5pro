<template>
  <div class="page">
    <div class="page-overlay"></div>
    <div class="child">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="title-section">
        <h1 class="title">
          <div class="uptime-value" style="font-size: 20px">
            已运行{{ formatUptime(deviceInfo.device_uptime as any) }}
          </div>
        </h1>
        <div class="status-indicator">
          <div :class="['status-dot', dataReady ? 'online' : 'offline']"></div>
          <span class="status-text">{{ dataReady ? '已连接' : '未加载' }}</span>
        </div>
      </div>

      <div class="controls">
        <div style="display: flex; gap: 10px">

          <div style="display: flex; position: relative">
            <span class="uptime-label" style="color: var(--theme-text-color, rgba(255,255,255,0.85))">{{ netWorkProvider }} {{ networkType }}{{ is5GA ? 'A' : '' }}</span>

          </div>

          <div style="display: flex; align-items: center">
            <span class="net-tag">{{ networkType }}{{ is5GA ? 'A' : '' }}</span>
          </div>
        </div>

        <div class="ctrl-group ctrl-row">
          <button
            class="btn adb-toggle"
            :class="autoRefresh ? 'btn-success' : 'btn-secondary'"
            @click="toggleAutoRefresh"
          >
            {{ autoRefresh ? '停止刷新' : '开始刷新' }}
          </button>
          <button
            class="btn adb-toggle"
            :class="adbStatus ? 'btn-disabled' : 'btn-primary'"
            :disabled="adbStatus"
            @click="handleOpenAdbClick"
          >开启 ADB</button>
          <button
            class="btn adb-toggle"
            :class="!adbStatus ? 'btn-disabled' : 'btn-danger'"
            :disabled="!adbStatus"
            @click="handleCloseAdbClick"
          >关闭 ADB</button>
          <button
            class="btn sys-action-toggle"
            :class="restarting ? 'btn-disabled' : 'btn-warning'"
            :disabled="restarting"
            @click="handleRestartClick"
          >重启设备</button>
          <button
            class="btn sys-action-toggle"
            :class="(firewallStatus !== false || firewallStatus === null) ? 'btn-disabled' : 'btn-primary'"
            :disabled="firewallStatus !== false"
            @click="handleOpenFirewallClick"
          >开启防火墙</button>
          <button
            class="btn sys-action-toggle"
            :class="firewallStatus !== true ? 'btn-disabled' : 'btn-danger'"
            :disabled="firewallStatus !== true"
            @click="handleCloseFirewallClick"
          >关闭防火墙</button>
        </div>

        <div v-if="wifiStatus?.main2g_ssid !== wifiStatus?.main5g_ssid" class="ctrl-group">
          2.4G-WIFI: {{wifiInfo.wifiStatus24?'开':'关'}}
          <button style="margin-left: 1px;" class="btn" :class="wifiInfo.wifiStatus24 ? 'btn-primary' : 'btn-primary'"
                  @click="wifiStateSetHandler('ra0',!wifiInfo.wifiStatus24)">{{wifiInfo.wifiStatus24?'关闭':'开启'}}</button>
          5G-WIFI: {{wifiInfo.wifiStatus5?'开':'关'}}
          <button style="margin-left: 1px;" class="btn" :class="wifiInfo.wifiStatus5 ? 'btn-primary' : 'btn-primary'"
                  @click="wifiStateSetHandler('rai0',!wifiInfo.wifiStatus5)">{{wifiInfo.wifiStatus5?'关闭':'开启'}}</button>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !dataReady" class="loading">
      <div class="loading-spinner"></div>
      <p>正在加载数据...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error">
      <div class="error-icon">⚠️</div>
      <h3>加载失败</h3>
      <p style="margin: 10px 0">{{ error }}</p>
      <button class="btn btn-danger" @click="refresh">重试</button>
    </div>

    <!-- 数据展示 -->
    <div v-else-if="dataReady" class="content">
      <div class="top-cards">
        <!-- NR 5G 信号卡片 -->
        <div class="card" v-if="networkType === '5G'">
          <div class="card-header">
            <h3 class="hd">
              <span class="hd-icon"><IconSignal5G /></span>NR 5G 信号
            </h3>
            <div class="card-tags">
              <span class="tag success">已激活</span>
              <span :class="['tag', getNetworkSignalStatus('nr').className]">
                信号{{ getNetworkSignalStatus('nr').text }}
              </span>
            </div>
          </div>
          <div class="card-content">
            <div class="signal-grid">
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSRP</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('nr', 'rsrp')"
                      @click="toggleSignalHelp('nr', 'rsrp')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('nr', 'rsrp', d.nr5g_rsrp).className]">
                    {{ getSignalDisplayStatus('nr', 'rsrp', d.nr5g_rsrp).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('nr', 'rsrp', d.nr5g_rsrp).className"
                    :style="{ width: getRsrpPercent(d.nr5g_rsrp) + '%' }"></div>
                  <span class="progress-text">{{ formatDbm(d.nr5g_rsrp) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('nr', 'rsrp')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rsrp.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rsrp.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rsrp.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSRQ</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('nr', 'rsrq')"
                      @click="toggleSignalHelp('nr', 'rsrq')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('nr', 'rsrq', d.nr5g_rsrq).className]">
                    {{ getSignalDisplayStatus('nr', 'rsrq', d.nr5g_rsrq).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('nr', 'rsrq', d.nr5g_rsrq).className"
                    :style="{ width: getRsrqPercent(d.nr5g_rsrq) + '%' }"></div>
                  <span class="progress-text">{{ formatDb(d.nr5g_rsrq) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('nr', 'rsrq')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rsrq.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rsrq.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rsrq.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">SINR</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('nr', 'sinr')"
                      @click="toggleSignalHelp('nr', 'sinr')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('nr', 'sinr', d.nr5g_snr).className]">
                    {{ getSignalDisplayStatus('nr', 'sinr', d.nr5g_snr).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('nr', 'sinr', d.nr5g_snr).className"
                    :style="{ width: getSnrPercent(d.nr5g_snr) + '%' }"></div>
                  <span class="progress-text">{{ formatSnr(d.nr5g_snr) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('nr', 'sinr')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.sinr.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.sinr.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.sinr.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSSI</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('nr', 'rssi')"
                      @click="toggleSignalHelp('nr', 'rssi')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('nr', 'rssi', d.nr5g_rssi).className]">
                    {{ getSignalDisplayStatus('nr', 'rssi', d.nr5g_rssi).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('nr', 'rssi', d.nr5g_rssi).className"
                    :style="{ width: getRssiPercent(d.nr5g_rssi) + '%' }"></div>
                  <span class="progress-text">{{ formatDbm(d.nr5g_rssi) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('nr', 'rssi')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rssi.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rssi.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rssi.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <span class="label">PCI</span>
                <span class="value">{{ d.nr5g_pci ?? '-' }}</span>
              </div>

              <div class="signal-item" width="100%">
                <span class="label">Cell ID</span>
                <span class="value">{{ d.nr5g_cell_id ?? '-' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- NR 5G 载波 -->
        <div class="card"  v-if="networkType === '5G'">
          <div class="card-header">
            <h3 class="hd">
              <span class="hd-icon"><IconCarrier5G /></span>5G 载波信息
            </h3>
            <span class="tag success">
              {{ networkType }}{{ is5GA ? 'A' : '' }}

              ({{ d.nr5g_action_band?.toUpperCase() ?? '-' }}{{
                  formatNrca(d.nrca,'',0,3) != '-' ? ', N' + formatNrca(d.nrca,'',0,3) : '' }}{{
                  formatNrca(d.nrca,'',1,3) != '-' ? ', N' + formatNrca(d.nrca,'',1,3) : '' }})
              </span>
          </div>
          <div class="card-content">
            <div class="signal-grid">
              <div :class="{ 'table-wrapper': themeStore.tableScrollEnabled }">
                <table class="mytable" width="100%">
                  <tr>
                    <td width="13%"></td>
                    <td width="9%">PCI</td>
                    <td width="11%">频段</td>
                    <td width="16%">频点</td>
                    <td width="11%">带宽</td>
                    <td width="10%">RSRP</td>
                    <td width="10%">RSRQ</td>
                    <td width="10%">SINR</td>
                    <td width="10%">RSSI</td>
                  </tr>
                  <tr>
                    <td>PCC</td>
                    <td>{{ d.nr5g_pci ?? '-' }}</td>
                    <td>{{ d.nr5g_action_band?.toUpperCase() ?? '-' }}</td>
                    <td>{{ d.nr5g_action_channel ?? '-' }}</td>
                    <td>{{ d.nr5g_bandwidth ? d.nr5g_bandwidth + 'Mhz' : '-' }}</td>
                    <td class="dbmstyle">{{ d.nr5g_rsrp }}</td>
                    <td>{{ d.nr5g_rsrq }}</td>
                    <td>{{ d.nr5g_snr }}</td>
                    <td class="dbmstyle">{{ d.nr5g_rssi }}</td>
                  </tr>
                  <tr>
                    <td>SCC0</td>
                    <td>{{ formatNrca(d.nrca,'',0,1) }}</td>
                    <td>{{ formatNrca(d.nrca,'N',0,3) }}</td>
                    <td>{{ formatNrca(d.nrca,'',0,4) }}</td>
                    <td>
                      {{
                        formatNrca(d.nrca, '', 0, 5) != '-'
                          ? formatNrca(d.nrca, '', 0, 5) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',0,7) }}</td>
                    <td>{{ formatNrca(d.nrca,'',0,8) }}</td>
                    <td>{{ formatNrca(d.nrca,'',0,9) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',0,10) }}</td>
                  </tr>
                  <tr>
                    <td>SCC1</td>
                    <td>{{ formatNrca(d.nrca,'',1,1) }}</td>
                    <td>{{ formatNrca(d.nrca,'N',1,3) }}</td>
                    <td>{{ formatNrca(d.nrca,'',1,4) }}</td>
                    <td>
                      {{
                        formatNrca(d.nrca, '', 1, 5) != '-'
                          ? formatNrca(d.nrca, '', 1, 5) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',1,7) }}</td>
                    <td>{{ formatNrca(d.nrca,'',1,8) }}</td>
                    <td>{{ formatNrca(d.nrca,'',1,9) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',1,10) }}</td>
                  </tr>
                  <tr>
                    <td>SCC2</td>
                    <td>{{ formatNrca(d.nrca,'',2,1) }}</td>
                    <td>{{ formatNrca(d.nrca,'N',2,3) }}</td>
                    <td>{{ formatNrca(d.nrca,'',2,4) }}</td>
                    <td>
                      {{
                        formatNrca(d.nrca, '', 2, 5) != '-'
                          ? formatNrca(d.nrca, '', 2, 5) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',2,7) }}</td>
                    <td>{{ formatNrca(d.nrca,'',2,8) }}</td>
                    <td>{{ formatNrca(d.nrca,'',2,9) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.nrca,'',2,10) }}</td>
                  </tr>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- LTE 4G 信号卡片 -->
        <div class="card" v-if="networkType === '4G'">
          <div class="card-header">
            <h3 class="hd">
              <span class="hd-icon"><IconSignalLte /></span>LTE 信号
            </h3>
            <div class="card-tags">
              <span class="tag success">已激活</span>
              <span :class="['tag', getNetworkSignalStatus('lte').className]">
                信号{{ getNetworkSignalStatus('lte').text }}
              </span>
            </div>
          </div>
          <div class="card-content">
            <div class="signal-grid">

              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSRP</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('lte', 'rsrp')"
                      @click="toggleSignalHelp('lte', 'rsrp')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('lte', 'rsrp', d.lte_rsrp).className]">
                    {{ getSignalDisplayStatus('lte', 'rsrp', d.lte_rsrp).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('lte', 'rsrp', d.lte_rsrp).className"
                    :style="{ width: getRsrpPercent(d.lte_rsrp) + '%' }"></div>
                  <span class="progress-text">{{ formatDbm(d.lte_rsrp) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('lte', 'rsrp')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rsrp.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rsrp.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rsrp.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSRQ</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('lte', 'rsrq')"
                      @click="toggleSignalHelp('lte', 'rsrq')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('lte', 'rsrq', d.lte_rsrq).className]">
                    {{ getSignalDisplayStatus('lte', 'rsrq', d.lte_rsrq).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('lte', 'rsrq', d.lte_rsrq).className"
                    :style="{ width: getRsrqPercent(d.lte_rsrq) + '%' }"></div>
                  <span class="progress-text">{{ formatDb(d.lte_rsrq) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('lte', 'rsrq')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rsrq.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rsrq.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rsrq.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">SINR</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('lte', 'sinr')"
                      @click="toggleSignalHelp('lte', 'sinr')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('lte', 'sinr', d.lte_snr).className]">
                    {{ getSignalDisplayStatus('lte', 'sinr', d.lte_snr).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('lte', 'sinr', d.lte_snr).className"
                    :style="{ width: getSnrPercent(d.lte_snr) + '%' }"></div>
                  <span class="progress-text">{{ formatSnr(d.lte_snr) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('lte', 'sinr')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.sinr.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.sinr.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.sinr.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <div class="signal-label-row">
                  <span class="signal-label-help">
                    <span class="label">RSSI</span>
                    <button
                      type="button"
                      class="signal-help-trigger"
                      :aria-expanded="isSignalHelpOpen('lte', 'rssi')"
                      @click="toggleSignalHelp('lte', 'rssi')">*</button>
                  </span>
                  <span :class="['signal-status', getSignalDisplayStatus('lte', 'rssi', d.lte_rssi).className]">
                    {{ getSignalDisplayStatus('lte', 'rssi', d.lte_rssi).text }}
                  </span>
                </div>
                <div class="progress-bar">
                  <div
                    class="progress-fill"
                    :class="getSignalDisplayStatus('lte', 'rssi', d.lte_rssi).className"
                    :style="{ width: getRssiPercent(d.lte_rssi) + '%' }"></div>
                  <span class="progress-text">{{ formatDbm(d.lte_rssi) }}</span>
                </div>
                <div v-if="isSignalHelpOpen('lte', 'rssi')" class="signal-help-panel">
                  <div class="signal-help-title">{{ signalHelpMap.rssi.title }}</div>
                  <div class="signal-help-desc">{{ signalHelpMap.rssi.description }}</div>
                  <div class="signal-help-ranges">
                    <div v-for="item in signalHelpMap.rssi.ranges" :key="item.label">
                      <span :class="['signal-help-dot', item.className]"></span>
                      <span>{{ item.label }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="signal-item">
                <span class="label">PCI</span>
                <span class="value">{{ d.lte_pci ?? '-' }}</span>
              </div>
              <div class="signal-item">
                <span class="label">Cell ID</span>
                <span class="value">{{ d.cell_id ?? '-' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- LTE 4G 载波 -->
        <div class="card" v-if="networkType === '4G'">
          <div class="card-header">
            <h3 class="hd">
              <span class="hd-icon"><IconCarrier4G /></span>4G 载波信息
            </h3>
            <span class="tag success">
              {{ networkType }}{{ is5GA ? '+' : '' }}
              ({{ formatNrca(d.lteca,'B',0,1) }}{{
                  formatNrca(d.lteca,'',1,1) != '-' ? ', B' + formatNrca(d.lteca,'',1,1) : '' }}{{
                  formatNrca(d.lteca,'',2,1) != '-' ? ', B' + formatNrca(d.lteca,'',2,1) : '' }}{{
                  formatNrca(d.lteca,'',3,1) != '-' ? ', B' + formatNrca(d.lteca,'',3,1) : '' }})
            </span>
          </div>
          <div class="card-content">
            <div class="signal-grid">
              <div :class="{ 'table-wrapper': themeStore.tableScrollEnabled }">
                <table class="mytable" width="100%">
                  <tr>
                    <td width="13%"></td>
                    <td width="9%">PCI</td>
                    <td width="11%">频段</td>
                    <td width="16%">信道</td>
                    <td width="11%">带宽</td>
                    <td width="10%">RSRP</td>
                    <td width="10%">RSRQ</td>
                    <td width="10%">SINR</td>
                    <td width="10%">RSSI</td>
                  </tr>
                  <tr>
                    <td>PCC</td>
                    <td>{{ d.lte_pci ?? '-' }}</td>
                    <td>{{ formatNrca(d.lteca,'B',0,1) }}</td>
                    <td>{{ d.wan_active_channel ?? '-' }}</td>
                    <td>
                      {{
                        formatNrca(d.lteca, '', 0, 4) != '-'
                          ? formatNrca(d.lteca, '', 0, 4) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ d.lte_rsrp }}</td>
                    <td>{{ d.lte_rsrq }}</td>
                    <td>{{ d.lte_snr }}</td>
                    <td class="dbmstyle">{{ d.lte_rssi }}</td>
                  </tr>
                  <tr>
                    <td>SCC0</td>
                    <td>{{ formatNrca(d.lteca,'',1,0) }}</td>
                    <td>{{ formatNrca(d.lteca,'B',1,1) }}</td>
                    <td>{{ formatNrca(d.lteca,'',1,3) }}</td>
                    <td>{{
                        formatNrca(d.lteca, '', 1, 4) != '-'
                          ? formatNrca(d.lteca, '', 1, 4) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',0,0) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',0,1) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',0,2) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',0,3) }}</td>
                  </tr>
                  <tr>
                    <td>SCC1</td>
                    <td>{{ formatNrca(d.lteca,'',2,0) }}</td>
                    <td>{{ formatNrca(d.lteca,'B',2,1) }}</td>
                    <td>{{ formatNrca(d.lteca,'',2,3) }}</td>
                    <td>{{
                        formatNrca(d.lteca, '', 2, 4) != '-'
                          ? formatNrca(d.lteca, '', 2, 4) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',1,0) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',1,1) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',1,2) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',1,3) }}</td>
                  </tr>
                  <tr>
                    <td>SCC2</td>
                    <td>{{ formatNrca(d.lteca,'',3,0) }}</td>
                    <td>{{ formatNrca(d.lteca,'B',3,1) }}</td>
                    <td>{{ formatNrca(d.lteca,'',3,3) }}</td>
                    <td>{{
                        formatNrca(d.lteca, '', 3, 4) != '-'
                          ? formatNrca(d.lteca, '', 3, 4) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',2,0) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',2,1) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',2,2) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',2,3) }}</td>
                  </tr>
                  <tr>
                    <td>SCC3</td>
                    <td>{{ formatNrca(d.lteca,'',4,0) }}</td>
                    <td>{{ formatNrca(d.lteca,'B',4,1) }}</td>
                    <td>{{ formatNrca(d.lteca,'',4,3) }}</td>
                    <td>{{
                        formatNrca(d.lteca, '', 4, 4) != '-'
                          ? formatNrca(d.lteca, '', 4, 4) + 'Mhz'
                          : '-'
                      }}
                    </td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',4,0) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',4,1) }}</td>
                    <td>{{ formatNrca(d.ltecasig,'',4,2) }}</td>
                    <td class="dbmstyle">{{ formatNrca(d.ltecasig,'',4,3) }}</td>
                  </tr>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
      

      <!-- 设备信息卡片 -->
      <div class="card device-info-card">
        <div class="card-header">
          <h3 class="hd">
            <span class="hd-icon"><IconDevice /></span>设备信息
          </h3>
        </div>
        <div class="card-content">
          <div class="device-stats">

            <div class="device-item">
              <div class="health-card cpu-health-card">
                <div class="health-title">CPU 负载</div>

                <div class="cpu-health-layout">
                  <div class="cpu-pie-box">
                    <div class="cpu-pie" :style="cpuPieStyle">
                      <div class="cpu-pie-inner">
                        <div class="cpu-pie-value">{{ totalCpuLoad.toFixed(0) }}%</div>
                      </div>
                    </div>
                  </div>

                  <div class="cpu-core-grid">
                    <div
                      class="cpu-core-card"
                      v-for="item in cpuCoreLoads"
                      :key="item.name"
                    >
                      <div class="cpu-core-header">
                        <!-- <span class="cpu-core-name">{{ item.name }}</span> -->
                        <span class="cpu-core-value">{{ item.value.toFixed(0) }}%</span>
                      </div>

                      <div class="cpu-core-bar">
                        <div
                          class="cpu-core-fill"
                          :style="{ width: item.value + '%' }"
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>


            <div class="device-item">
              <div class="health-card temp-health-card">
                <div class="health-title">温度状态</div>

                <div class="temp-gauges">
                  <div class="temp-gauge">
                    <div
                      class="temp-ring"
                      :class="getTempClass(cpuTemp.cpuss_temp)"
                      :style="{ '--percent': getTempPercent(cpuTemp.cpuss_temp) + '%' }"
                    >
                      <div class="temp-ring-inner">
                        <strong>{{ cpuTemp.cpuss_temp || '-' }}°</strong>
                        <span>CPU</span>
                      </div>
                    </div>
                    <div class="temp-state" :class="getTempClass(cpuTemp.cpuss_temp)">
                      {{ getTempText(cpuTemp.cpuss_temp) }}
                    </div>
                  </div>

                  <div class="temp-gauge">
                    <div
                      class="temp-ring"
                      :class="getTempClass(thermal.modem_temp)"
                      :style="{ '--percent': getTempPercent(thermal.modem_temp) + '%' }"
                    >
                      <div class="temp-ring-inner">
                        <strong>{{ thermal.modem_temp || '-' }}°</strong>
                        <span>基带</span>
                      </div>
                    </div>
                    <div class="temp-state" :class="getTempClass(thermal.modem_temp)">
                      {{ getTempText(thermal.modem_temp) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="device-item">
              <div class="health-card memory-health-card">
                <div class="health-title">内存使用</div>

                <div class="memory-big">
                  {{
                    formatMemoryPercent(
                      ((deviceInfo.meminfo?.total as any) || 0) -
                        ((deviceInfo.meminfo?.avaliable as any) || 0),
                      (deviceInfo.meminfo?.total as any) || 1
                    )
                  }}%
                </div>

                <div class="memory-detail">
                  {{
                    formatMemory(
                      ((deviceInfo.meminfo?.total || 0) as any) -
                        ((deviceInfo.meminfo?.avaliable as any) || 0)
                    )
                  }}
                  <span>/ {{ formatMemory((deviceInfo.meminfo?.total as any) || 0) }}</span>
                </div>

                <div class="memory-stack">
                  <div
                    class="memory-stack-fill"
                    :style="{
                      width:
                        formatMemoryPercent(
                          ((deviceInfo.meminfo?.total as any) || 0) -
                            ((deviceInfo.meminfo?.avaliable as any) || 0),
                          (deviceInfo.meminfo?.total as any) || 1
                        ) + '%',
                    }"
                  ></div>
                </div>

                <div class="memory-caption">已用 / 总量</div>
              </div>
            </div>

            <!-- <div class="device-item">
              <div class="device-label">
                CPU 负载
                {{
                  (100 - (deviceInfo.cpuinfo?.[0]?.idle as any) || 0).toFixed(2)
                }}
                %
              </div>
              <div class="device-values">
                <div class="load-item">
                  <span class="load-label">核心1</span>
                  <span class="load-value"
                    >{{
                      (
                        100 - (deviceInfo.cpuinfo?.[1]?.idle as any) || 0
                      ).toFixed(0)
                    }}
                    %</span
                  >
                </div>
                <div class="load-item">
                  <span class="load-label">核心2</span>
                  <span class="load-value"
                    >{{
                      (
                        100 - (deviceInfo.cpuinfo?.[2]?.idle as any) || 0
                      ).toFixed(0)
                    }}
                    %</span
                  >
                </div>
                <div class="load-item">
                  <span class="load-label">核心3</span>
                  <span class="load-value"
                    >{{
                      (
                        100 - (deviceInfo.cpuinfo?.[3]?.idle as any) || 0
                      ).toFixed(0)
                    }}
                    %</span
                  >
                </div>
                <div class="load-item">
                  <span class="load-label">核心4</span>
                  <span class="load-value"
                    >{{
                      (
                        100 - (deviceInfo.cpuinfo?.[4]?.idle as any) || 0
                      ).toFixed(0)
                    }}
                    %</span
                  >
                </div>
              </div>
            </div> -->
          </div>
        </div>
      </div>

      <!-- 网络信息 / 流量统计 / 接口状态：三卡片等宽网格 -->
      <div class="bottom-cards">
      <!-- 网络信息卡片 -->
      <div class="card">
        <div class="card-header">
          <h3 class="hd">
            <span class="hd-icon"><IconInternet /></span>网络信息
          </h3>
        </div>
        <div class="card-content">
          <div class="ambr-hero">
            <div class="ambr-tile down">
              <span class="ambr-ico"><IconArrowDown /></span>
              <div class="ambr-meta">
                <div class="ambr-label">下行速率</div>
                <div class="ambr-val">{{ netAmbr.dl.value }}<span class="ambr-unit">{{ formatSpeedUnit(netAmbr.dl.unit) }}</span></div>
              </div>
            </div>
            <div class="ambr-tile up">
              <span class="ambr-ico"><IconArrowUp /></span>
              <div class="ambr-meta">
                <div class="ambr-label">上行速率</div>
                <div class="ambr-val">{{ netAmbr.ul.value }}<span class="ambr-unit">{{ formatSpeedUnit(netAmbr.ul.unit) }}</span></div>
              </div>
            </div>
            <div class="ambr-qci">
              <div class="ambr-label">QCI</div>
              <div class="ambr-qci-val">{{ netAmbr.qci2 || netAmbr.qci1 }}</div>
              <div class="ambr-qci-sub" v-if="qciHint">{{ qciHint }}</div>
            </div>
            <div class="ambr-signal" :class="networkScoreTileClass">
              <div class="ambr-label">网络综合评分</div>
              <div class="ambr-score-value" :style="{ color: networkScoreColor }">{{ networkScore }}</div>
              <!-- 简化：4 子指标（RSRP/RSRQ/SINR/RSSI）已隐藏，仅保留主分数 -->
              <!-- <div class="ambr-score-detail">
                <span><em :style="{ color: networkScoreColor }">{{ fmtInt(activeRsrp) }}</em><i>RSRP</i></span>
                <span><em :style="{ color: networkScoreColor }">{{ fmtInt(activeRsrq) }}</em><i>RSRQ</i></span>
                <span><em :style="{ color: networkScoreColor }">{{ formatSnrNoUnit(activeSinr) }}</em><i>SINR</i></span>
                <span><em :style="{ color: networkScoreColor }">{{ fmtInt(activeRssi) }}</em><i>RSSI</i></span>
              </div> -->
            </div>
          </div>

          <div class="ct-divider"></div>
          <div class="info-section-title info-section-title-row">
            <span class="info-section-title-main">RF 频段功率</span>
            <span class="info-section-title-hint">控制 CPE modem RF 发射功率上限</span>
          </div>
          <div class="ct-rf-bar">
            <el-button
              size="small"
              class="sys-action-btn is-amber"
              title="向 modem 写入各频段最大发射功率（约 29dBm），NSA + SA 常用频段全覆盖"
              :loading="rfSetLoading"
              :disabled="rfSetLoading || rfStatus?.state === 'applied_target'"
              @click="onSetRfMaxpower"
            >
              <span class="btn-ico">⚡</span>{{ rfStatus?.state === 'applied_target' ? '已是 29dBm（已锁定）' : '所有频段功率设为 29dBm' }}
            </el-button>
            <el-button
              size="small"
              class="sys-action-btn is-slate"
              title="恢复 modem 各频段发射功率出厂默认值"
              :loading="rfDefaultLoading"
              :disabled="rfDefaultLoading || rfStatus?.state === 'default'"
              @click="onSetRfDefault"
            >
              <span class="btn-ico">↺</span>{{ rfStatus?.state === 'default' ? '已是默认（已锁定）' : '恢复默认功率' }}
            </el-button>
            <!-- 状态 tag 兼做"当前是什么"的可见标识 -->
            <span
              class="rf-state-tag"
              :class="rfStatus?.state === 'applied_target' ? 'on' : (rfStatus?.state === 'custom' ? 'warn' : (rfStatus?.state === 'unknown' ? 'warn' : 'off'))"
              v-if="rfStatus"
            >
              {{ rfStateText }}
            </span>
          </div>
          <div class="rf-drift-warn" v-if="rfStatus && !rfStatus.consistent && rfStatus.drift_msg">
            <span class="warn-ico">⚠</span>{{ rfStatus.drift_msg }}
          </div>
          <div class="rf-note">状态反映 modem 寄存器写入的指令；实际发射功率由基站调度，本工具无法强制。</div>

          <div class="info-grid">
            <div class="info-section-title">SIM / 网络</div>
            <div class="info-item">
              <span class="label">运营商</span>
              <span class="value">{{ netWorkProvider }}</span>
            </div>
            <div class="info-item">
              <span class="label">网络类型</span>
              <span class="value">{{ networkType }}{{ is5GA ? 'A' : '' }}</span>
            </div>
            <!-- <div class="info-item">
              <span class="label">驻网状态</span>
              <span class="value">{{ d.simcard_roam || '-' }}</span>
            </div> -->
            <!-- <div class="info-item">
              <span class="label">选择模式</span>
              <span class="value">{{ d.net_select_mode || '-' }}</span>
            </div> -->
            <!-- <div class="info-item">
              <span class="label">选择策略</span>
              <span class="value">{{ d.net_select || '-' }}</span>
            </div> -->
            <div class="info-item">
              <span class="label">信号强度</span>
              <div class="signal-strength-row">
                <span class="value signal-strength-num">{{ activeRsrp != null && Number.isFinite(activeRsrp) ? activeRsrp + ' dBm' : '—' }}</span>
                <div class="signal-bars" :title="`信号 ${signalBars}/5`">
                  <div
                    v-for="n in 5"
                    :key="n"
                    :class="['bar', { active: n <= signalBars }]"></div>
                </div>
              </div>
            </div>
            <div class="info-item">
              <span class="label">连接数量</span>
              <span class="value"
                >有线：{{ lanUserList?.lan_num || '0' }} / 无线：{{
                  lanUserList?.wireless_num || '0'
                }}</span
              >
            </div>
            <div class="info-section-title">频段 / 频道</div>
            <div class="info-item">
              <span class="label">主载波</span>
              <span class="value">{{
                d.wan_active_band?.toUpperCase() || '-'
              }}</span>
            </div>
            <div class="info-item">
              <span class="label">当前载波</span>
              <span class="value"
                >{{
                  d.wan_active_band?.toUpperCase()
                    ? d.wan_active_band.toUpperCase() + ', '
                    : ''
                }}{{ currentActiveBands || '-' }}</span
              >
            </div>
            <div class="info-item">
              <span class="label">频道</span>
              <span class="value">{{ d.nr5g_action_channel ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">带宽</span>
              <span class="value">{{
                d.nr5g_bandwidth ? d.nr5g_bandwidth + ' Mhz' : '-'
              }}</span>
            </div>
            <!-- <div class="info-item">
              <span class="label">LTE 锁频</span>
              <span class="value">{{ d.lte_band_lock || '-' }}</span>
            </div> -->
            <!-- <div class="info-item">
              <span class="label">NR SA 锁频</span>
              <span class="value">{{ d.nr5g_sa_band_lock || '-' }}</span>
            </div> -->
            <!-- <div class="info-item">
              <span class="label">LTE 频段</span>
              <span
                class="value"
                style="
                  white-space: pre-wrap;
                  word-wrap: break-word;
                  overflow: hidden;
                "
                >{{ d.lte_band || '-' }}</span
              >
            </div> -->
            <div class="info-section-title">设备标识</div>
            <div class="info-item">
              <span class="label">SIM 卡号</span>
              <span class="value">{{ netAmbr.sim_number || simInfo2?.msisdn || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">IMSI</span>
              <span class="value">{{ simInfo2.sim_imsi ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">IMEI</span>
              <span class="value">{{ simInfo?.values?.imei ?? '-' }}</span>
            </div>
            <!-- <div class="info-item">
              <span class="label">Lock Status</span>
              <span class="value">{{
                simInfo?.values?.lock_status ?? '-'
              }}</span>
            </div> -->
            <div class="info-item">
              <span class="label">Modem MSN</span>
              <span class="value">{{ simInfo?.values?.modem_msn ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">WLAN MAC</span>
              <span class="value">{{
                simInfo?.values?.wlan_mac_address ?? '-'
              }}</span>
            </div>
            <div class="info-item">
              <span class="label">系统版本</span>
              <span class="value">{{
                sysVersion?.wa_inner_version ?? '-'
              }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 流量统计卡片 -->
      <div class="card">
        <div class="card-header">
          <h3 class="hd">
            <span class="hd-icon"><IconChart /></span>流量统计
          </h3>
        </div>
        <div class="card-content">
          <div class="traffic-stats">
            <div class="traffic-item">
              <div class="traffic-label">上传速度</div>
              <div class="traffic-value upload">
                {{ formatSpeed(trafficData.real_tx_speed) }}
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">下载速度</div>
              <div class="traffic-value download">
                {{ formatSpeed(trafficData.real_rx_speed) }}
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">上传用量</div>
              <div class="traffic-value traffic-value-stack">
                <!-- 单行紧凑：今日 · 本月；切分符竖线分隔，与速率对齐 -->
                <div class="traffic-stack-row">
                  <span class="traffic-stack-k">今日</span>
                  <span class="traffic-stack-v">{{ formatBytes(trafficData.day_tx_bytes) }}</span>
                </div>
                <div class="traffic-stack-row">
                  <span class="traffic-stack-k">本月</span>
                  <span class="traffic-stack-v">{{ formatBytes(trafficData.month_tx_bytes) }}</span>
                </div>
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">下载用量</div>
              <div class="traffic-value traffic-value-stack">
                <div class="traffic-stack-row">
                  <span class="traffic-stack-k">今日</span>
                  <span class="traffic-stack-v">{{ formatBytes(trafficData.day_rx_bytes) }}</span>
                </div>
                <div class="traffic-stack-row">
                  <span class="traffic-stack-k">本月</span>
                  <span class="traffic-stack-v">{{ formatBytes(trafficData.month_rx_bytes) }}</span>
                </div>
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">总上传</div>
              <div class="traffic-value">
                {{ formatBytes(trafficData.total_tx_bytes as any) }}
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">总下载</div>
              <div class="traffic-value">
                {{ formatBytes(trafficData.total_rx_bytes as any) }}
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">最大上传速度</div>
              <div class="traffic-value">
                {{ formatSpeed(trafficData.real_max_tx_speed) }}
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-label">最大下载速度</div>
              <div class="traffic-value">
                {{ formatSpeed(trafficData.real_max_rx_speed) }}
              </div>
            </div>
          </div>

          <!-- conntrack-tune 装机/卸载子区块（位于流量统计卡片内） -->
          <div class="conn-track-block">
            <div class="ct-divider"></div>
            <div class="ct-header">
              <span class="hd-icon"><IconLink /></span>
              <span class="ct-title">连接状态 · conntrack-tune</span>
            </div>
            <div class="ct-actions">
              <el-button
                size="small"
                class="sys-action-btn is-green"
                title="写入 /etc/init.d/conntrack-tune 并开机自启：扩容 conntrack 表 + 缩短 UDP 超时，根治 PCDN 高并发丢包"
                :loading="ctInstallLoading"
                :disabled="ctInstallLoading || ctStatus?.installed"
                @click="onInstallConntrack"
              >
                <span class="btn-ico">✓</span>{{ ctStatus?.installed ? 'conntrack-tune 已启用' : '安装conntrack优化' }}
              </el-button>
              <el-button
                size="small"
                class="sys-action-btn is-red"
                title="禁用并删除 conntrack-tune 脚本（重启后失效，再次恢复需重装；自动清理旧名 zz-conntrack 残留）"
                :loading="ctUninstallLoading"
                :disabled="ctUninstallLoading || !ctStatus?.installed"
                @click="onUninstallConntrack"
              >
                <span class="btn-ico">✕</span>卸载conntrack优化
              </el-button>
            </div>

            <div class="ct-main">
              <div class="ct-count">
                <span class="ct-count-num">{{ connTrack?.count ?? '-' }}</span>
                <span class="ct-count-label">当前连接数</span>
              </div>
              <div class="ct-usage">
                <div class="ct-usage-bar">
                  <div
                    class="ct-usage-fill"
                    :class="ctUsageClass"
                    :style="{ width: (connTrack?.usage_percent ?? 0) + '%' }"
                  ></div>
                </div>
                <div class="ct-usage-text">
                  占用 {{ connTrack?.usage_percent ?? 0 }}% · 上限 {{ fmtInt(connTrack?.max ?? 0) }}
                </div>
              </div>
            </div>

            <div class="ct-section-title">协议分布</div>
            <div class="ct-proto" v-if="connTrack && connTrack.protocols && Object.keys(connTrack.protocols).length">
              <span class="ct-proto-chip" v-for="(cnt, p) in connTrack.protocols" :key="p">
                {{ p.toUpperCase() }} <b>{{ fmtInt(cnt) }}</b>
              </span>
            </div>
            <div class="ct-proto-empty" v-else>—</div>

            <div class="ct-section-title">占用最高设备</div>
            <div class="ct-top" v-if="connTrack?.top_sources?.length">
              <div
                class="ct-top-item"
                v-for="(s, i) in connTrack.top_sources"
                :key="s.ip"
                :class="{ 'is-top': i === 0 }"
              >
                <span class="ct-rank">{{ i + 1 }}</span>
                <span class="ct-name">{{ s.name }}</span>
                <span class="ct-ip">{{ s.ip }}</span>
                <span class="ct-cnt">{{ fmtInt(s.count) }}</span>
              </div>
              <div class="ct-top-item is-nonlan" v-if="connTrack.non_lan_count">
                <span class="ct-rank">—</span>
                <span class="ct-name" title="CPE 自身上行 / 外网连接（不在 LAN 设备表里的源，如运营商 IPv6、链路本地等）">
                  CPE 自身上行/外网
                </span>
                <span class="ct-ip">—</span>
                <span class="ct-cnt">{{ fmtInt(connTrack.non_lan_count) }}</span>
              </div>
            </div>
            <div class="ct-top-empty" v-else>
              暂无 LAN 设备连接
              <span v-if="connTrack?.non_lan_count" class="ct-top-empty-sub">
                （CPE 自身上行/外网 {{ fmtInt(connTrack.non_lan_count) }} 条）
              </span>
            </div>
          </div>
        </div>
      </div>

            <!-- 接口状态卡片 -->
      <div class="card">
        <div class="card-header">
          <h3 class="hd">
            <span class="hd-icon"><IconInterface /></span>接口状态
          </h3>
        </div>
        <div class="card-content">
          <div class="interface-grid">

            <div class="interface-section" v-if="wwanInfo?.ipv4_address">
              <h4>WAN IPv4</h4>
              <div class="info-grid-compact">
                <div class="info-item">
                  <span class="label">IP 地址</span>
                  <span class="value">{{ wwanInfo?.ipv4_address || '-' }}</span>
                </div>
                <div class="info-item">
                  <span class="label">网关</span>
                  <span class="value">{{ wwanInfo?.ipv4_gateway || '-' }}</span>
                </div>
                <div class="info-item">
                  <span class="label">DNS 服务器</span>
                  <span class="value">{{
                    wanData['dns-server']?.join('\n') || '-'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="label">运行时间</span>
                  <span class="value">{{ formatUptime(wanData.uptime) }}</span>
                </div>
              </div>
            </div>

            <div class="interface-section" v-if="wwanInfo?.ipv6_address !== '0'">
              <h4>WAN IPv6</h4>
              <div class="info-grid-compact">
                <div class="info-item">
                  <span class="label">IPv6 地址</span>
                  <span class="value">{{ wwanInfo?.ipv6_address || '-' }}</span>
                </div>
                <div class="info-item">
                  <span class="label">网关</span>
                  <span class="value">{{ wwanInfo?.ipv6_gateway || '-' }}</span>
                </div>
                <div class="info-item">
                  <span class="label">DNS 服务器</span>
                  <span class="value">{{
                    wan6Data['dns-server']?.join('\n') || '-'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="label">运行时间</span>
                  <span class="value">{{ formatUptime(wan6Data.uptime) }}</span>
                </div>
              </div>
            </div>
            
            <div class="interface-section" v-if="lanData?.ipv4_address">
              <h4>LAN 接口</h4>
              <div class="info-grid-compact">
                <div class="info-item">
                  <span class="label">IP 地址</span>
                  <span class="value">{{
                    lanData.ipv4_address?.[0]?.address || '-'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="label">网关</span>
                  <span class="value">{{
                    lanData.route?.[0]?.nexthop || '-'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="label">DNS 服务器</span>
                  <span class="value">{{
                    lanData['dns-server']?.join('\n') || '-'
                  }}</span>
                </div>
                <div class="info-item">
                  <span class="label">运行时间</span>
                  <span class="value">{{ formatUptime(lanData.uptime) }}</span>
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
      </div>

      <!-- 频段与锁定卡片 -->
      <!-- <div class="card">
        <div class="card-header">
          <h3 class="hd">
            <img style="width: 24px" :src="LockIcon" alt="" />频段与锁定
          </h3>
        </div>
        <div class="card-content">
          <div class="info-grid">
            <div class="info-item">
              <span class="label">主载波</span>
              <span class="value">{{
                d.wan_active_band?.toUpperCase() || '-'
              }}</span>
            </div>
            <div class="info-item">
              <span class="label">工作频段</span>
              <span class="value"
                >{{
                  d.wan_active_band?.toUpperCase()
                    ? d.wan_active_band.toUpperCase() + ', '
                    : ''
                }}{{ currentActiveBands || '-' }}</span
              >
            </div>
            <div class="info-item">
              <span class="label">频道</span>
              <span class="value">{{ d.nr5g_action_channel ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">带宽</span>
              <span class="value">{{
                d.nr5g_bandwidth ? d.nr5g_bandwidth + ' Mhz' : '-'
              }}</span>
            </div>
            <div class="info-item">
              <span class="label">LTE 锁频</span>
              <span class="value">{{ d.lte_band_lock || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">NR SA 锁频</span>
              <span class="value">{{ d.nr5g_sa_band_lock || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">LTE 频段</span>
              <span
                class="value"
                style="
                  white-space: pre-wrap;
                  word-wrap: break-word;
                  overflow: hidden;
                "
                >{{ d.lte_band || '-' }}</span
              >
            </div>
          </div>
        </div>
      </div> -->

      <!-- 标识信息卡片 -->
      <!-- <div class="card">
        <div class="card-header">
          <h3 class="hd">
            <img style="width: 24px" :src="TagIcon" alt="" />标识信息
          </h3>
        </div>
        <div class="card-content">
          <div class="info-grid">
            <div class="info-item">
              <span class="label">IMSI</span>
              <span class="value">{{ simInfo2.sim_imsi ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">IMEI</span>
              <span class="value">{{ simInfo?.values?.imei ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">Lock Status</span>
              <span class="value">{{
                simInfo?.values?.lock_status ?? '-'
              }}</span>
            </div>
            <div class="info-item">
              <span class="label">Modem MSN</span>
              <span class="value">{{ simInfo?.values?.modem_msn ?? '-' }}</span>
            </div>
            <div class="info-item">
              <span class="label">WLAN MAC</span>
              <span class="value">{{
                simInfo?.values?.wlan_mac_address ?? '-'
              }}</span>
            </div>
          </div>
        </div>
      </div> -->

    </div>

    <!-- 空状态 -->
    <div v-else class="empty">
      <div class="empty-icon">📱</div>
      <h3>暂无数据</h3>
      <p>请点击刷新按钮获取设备信息</p>
    </div>
    </div>

    <!-- 系统操作结果弹窗：装机/卸载/频段功率等动作的 backend details 展示 -->
    <el-dialog
      v-model="actionDialogVisible"
      :title="actionDialog.title"
      width="720px"
      append-to-body
      destroy-on-close
      class="action-result-dialog"
    >
      <!-- 顶部一行简短结果摘要 -->
      <div
        class="action-result-summary"
        :class="actionDialog.success ? 'is-ok' : 'is-fail'"
      >
        <span class="action-result-icon">{{ actionDialog.success ? '✓' : '✕' }}</span>
        <span class="action-result-msg">{{ actionDialog.summary }}</span>
      </div>

      <!-- 详细信息分组（按 section 渲染，每 section 一个折叠块的简化版） -->
      <div
        v-for="(sec, si) in actionDialog.sections"
        :key="si"
        class="action-result-section"
      >
        <div v-if="sec.title" class="action-result-section-title">{{ sec.title }}</div>
        <ul class="action-result-list">
          <li
            v-for="(item, ii) in sec.items"
            :key="ii"
            :class="['action-result-item', item.status || '']"
          >
            <span class="ar-label">{{ item.label }}</span>
            <span class="ar-value">
              <code v-if="item.code !== undefined">{{ item.code }}</code>
              <span v-if="item.value !== undefined && item.value !== ''" class="ar-pre">{{ item.value }}</span>
              <span v-else-if="item.value === ''" class="ar-empty">（空）</span>
            </span>
          </li>
        </ul>
      </div>

      <!-- 额外原始 details（如 AT 命令 echo） -->
      <div v-if="actionDialog.rawDetails && actionDialog.rawDetails.length" class="action-result-section">
        <div class="action-result-section-title">原始输出</div>
        <pre class="action-result-raw">{{ actionDialog.rawDetails.join('\n') }}</pre>
      </div>

      <template #footer>
        <el-button type="primary" @click="actionDialogVisible = false">关闭</el-button>
        <el-button v-if="actionDialog.rawDetails && actionDialog.rawDetails.length" @click="copyActionRaw">
          复制原始输出
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios';
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus';
import { computed, h, onMounted, onUnmounted, ref } from 'vue';
import { useThemeStore } from '@/stores/themeStore';

const themeStore = useThemeStore();

// 卡片头 / AMBR 图标：直接由 Vue 渲染函数产出真实 SVG 节点, 不依赖 v-html;
// 描边色继承父级 .hd-icon / .ambr-ico 的 color(currentColor), 在深色卡片上始终可见。
// 每个区块独立设计，辨识度高；.hd-icon 提供圆角徽标底，确保清晰可辨

const SVG_BASE = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  'stroke-width': 1.9,
  'stroke-linecap': 'round',
  'stroke-linejoin': 'round',
  'aria-hidden': 'true',
} as const;

function svg(children: any[]) {
  return h('svg', SVG_BASE, children);
}

// NR 5G 信号：信号强度柱
const IconSignal5G = () =>
  svg([
    h('path', { d: 'M3 20v-4' }),
    h('path', { d: 'M9 20v-8' }),
    h('path', { d: 'M15 20v-12' }),
    h('path', { d: 'M21 20V4' }),
  ]);
// 5G 载波信息：广播同心弧
const IconCarrier5G = () =>
  svg([
    h('circle', { cx: 12, cy: 12, r: 2.4 }),
    h('path', { d: 'M7.8 7.8a6.4 6.4 0 0 0 0 8.4' }),
    h('path', { d: 'M16.2 7.8a6.4 6.4 0 0 1 0 8.4' }),
    h('path', { d: 'M5 5a10.4 10.4 0 0 0 0 14' }),
    h('path', { d: 'M19 5a10.4 10.4 0 0 1 0 14' }),
  ]);
// LTE 信号：基站塔
const IconSignalLte = () =>
  svg([
    h('path', { d: 'M12 3v18' }),
    h('path', { d: 'M8 21h8' }),
    h('path', { d: 'M7 11l5-5 5 5' }),
    h('path', { d: 'M4.5 14.5l7.5-7.5 7.5 7.5' }),
  ]);
// 4G 载波信息：载波波形
const IconCarrier4G = () =>
  svg([
    h('path', { d: 'M2 12c2-4.2 4-4.2 6 0s4 4.2 6 0 4-4.2 6 0' }),
    h('path', { d: 'M2 17c2-4.2 4-4.2 6 0s4 4.2 6 0 4-4.2 6 0' }),
  ]);
// 设备信息：芯片
const IconDevice = () =>
  svg([
    h('rect', { x: 7, y: 7, width: 10, height: 10, rx: 2 }),
    h('path', { d: 'M10 7V4' }),
    h('path', { d: 'M14 7V4' }),
    h('path', { d: 'M10 20v-3' }),
    h('path', { d: 'M14 20v-3' }),
    h('path', { d: 'M7 10H4' }),
    h('path', { d: 'M7 14H4' }),
    h('path', { d: 'M20 10h-3' }),
    h('path', { d: 'M20 14h-3' }),
  ]);
// 网络信息：地球
const IconInternet = () =>
  svg([
    h('circle', { cx: 12, cy: 12, r: 9.2 }),
    h('path', { d: 'M2.8 12h18.4' }),
    h('path', {
      d:
        'M12 2.8c3 2.4 4.4 6 4.4 9.2s-1.4 6.8-4.4 9.2c-3-2.4-4.4-6-4.4-9.2S9 5.2 12 2.8z',
    }),
    h('path', { d: 'M5 8.5c2.8 1.8 11.2 1.8 14 0' }),
    h('path', { d: 'M5 15.5c2.8-1.8 11.2-1.8 14 0' }),
  ]);
// 流量统计：柱状图
const IconChart = () =>
  svg([
    h('path', { d: 'M4 20V11' }),
    h('path', { d: 'M10 20V4.5' }),
    h('path', { d: 'M16 20v-7' }),
    h('path', { d: 'M21.5 20V13' }),
  ]);
// 接口状态：端口插头
const IconInterface = () =>
  svg([
    h('path', { d: 'M9 2.5v6' }),
    h('path', { d: 'M15 2.5v6' }),
    h('path', { d: 'M7 8.5h10v3.5a5 5 0 0 1-10 0V8.5z' }),
    h('path', { d: 'M12 17v5' }),
  ]);
// 连接状态：网络节点
const IconLink = () =>
  svg([
    h('circle', { cx: 6, cy: 6, r: 2.4 }),
    h('circle', { cx: 18, cy: 6, r: 2.4 }),
    h('circle', { cx: 12, cy: 18, r: 2.4 }),
    h('path', { d: 'M7.8 7.6 11 16M16.2 7.6 13 16M8.4 6h7.2' }),
  ]);
// AMBR 上下行箭头
const IconArrowDown = () =>
  svg([h('path', { d: 'M12 4v11' }), h('path', { d: 'M6.5 11.5 12 17l5.5-5.5' })]);
const IconArrowUp = () =>
  svg([h('path', { d: 'M12 20V9' }), h('path', { d: 'M6.5 12.5 12 7l5.5 5.5' })]);


// interface UbusResponse<T = any> {
//   code: number;
//   msg: string;
//   result?: T;
// }

interface ZteRpcResponse {
  jsonrpc: string
  id: number
  result: [number, any]
}

interface NetInfoResult {
  [key: string]: any;
}

interface NetworkInterface {
  up: boolean;
  device?: string;
  proto?: string;
  uptime?: number;
  ipv4_address?: Array<{ address: string; mask: number }>;
  ipv6_address?: Array<{ address: string; mask: number }>;
  route?: Array<{ nexthop: string }>;
  'dns-server'?: string[];
}

interface TrafficData {
  real_tx_speed: number;
  real_rx_speed: number;
  real_tx_bytes: number;
  real_rx_bytes: number;
  real_max_tx_speed: number;
  real_max_rx_speed: number;
  total_tx_bytes: number;
  total_rx_bytes: number;
  day_tx_bytes: number;
  day_rx_bytes: number;
  month_tx_bytes: number;
  month_rx_bytes: number;
}

interface ConnTrackSource {
  ip: string;
  name: string;
  count: number;
}

interface ConnTrackData {
  count: number;
  max: number;
  usage_percent: number;
  protocols: Record<string, number>;
  top_sources: ConnTrackSource[];
  non_lan_count: number;
}

interface SystemInfo {
  localtime: number;
  uptime: number;
  load: number[];
  memory: {
    total: number;
    free: number;
    shared: number;
    buffered: number;
    available: number;
    cached: number;
  };
  root: {
    total: number;
    free: number;
    used: number;
    avail: number;
  };
  tmp: {
    total: number;
    free: number;
    used: number;
    avail: number;
  };
  swap: {
    total: number;
    free: number;
  };
}

interface DeviceInfo {
  hightemp_datalimit_status: string;
  quicken_power_on: string;
  bat_online: string;
  bat_health: string;
  bat_mode: string;
  bat_low_power: string;
  bat_percent: string;
  bat_level: string;
  bat_temperature: string;
  bat_charger_connect: string;
  bat_charger_type: string;
  bat_charger_status: string;
  bat_ui_charger_type: string;
  bat_temperature_level: string;
  external_charging_flag: string;
  bat_time_to_full: string;
  bat_time_to_empty: string;
  power_adapter: string;
  device_uptime: string;
  cpuinfo: {
    name: string;
    idle: string;
    gnice?: string;
  }[];
  meminfo: {
    total: string;
    free: string;
    avaliable: string;
  };
  flashinfo: {
    filesystem: string;
    size: string;
    used: string;
    avail: string;
    use: string;
    mounted_on: string;
  }[];
}

interface CpuTemp {
  cpuss_temp: number;
}

interface SimInfo {
  values: {
    digitalcode: string;
    imei: string;
    imei2: string;
    lock_status: string;
    modem_msn: string;
    wlan_mac_address: string;
  };
}

interface SimInfo2 {
  sim_iccid: string;
  sim_imsi: string;
  msisdn?: string;
  Operator: string;
}

interface WwanInfo {
  connect_fail_count: 0;
  connect_status: string;
  ipv4_address: string;
  ipv4_dev_name: string;
  ipv4_dns_prefer: string;
  ipv4_dns_standby: string;
  ipv4_gateway: string;
  ipv4_netmask: string;
  ipv6_address: string;
  ipv6_dev_name: string;
  ipv6_dns_prefer: string;
  ipv6_dns_standby: string;
  ipv6_gateway: string;
  roam_enable: number;
}

interface LanUserList {
  access_total_num: number;
  lan_num: number;
  wireless_num: number;
  offline_num: number;
  guest_num_24g: number;
  guest_num_5g: number;
  guest_num_6g: number;
}

// WiFi状态
interface WifiInfo {
  ra0: string
  rai0: string
  rax0: string
  wifiStatus24: boolean
  wifiStatus5: boolean
  highPerformance: boolean
}
const wifiInfo = ref<WifiInfo>({} as WifiInfo);
const wifiModeText = computed(() => wifiInfo.value.highPerformance ? '性能模式' : '省电模式')
const wifiButtonText = computed(() => wifiInfo.value.highPerformance ? '切换为省电' : '切换高性能')

interface WifiStatus {
  dfs_status: string,
  lbd_enable: string,
  load_status: string,
  main2g_authmode: string,
  main2g_ssid: string,
  main5g_authmode: string,
  main5g_ssid: string,
  mesh_deployed: string,
  mesh_deploying_status: string,
  mesh_set_status: string,
  mlo_enable: string,
  radio2: string,
  radio2_disabled: string,
  radio5: string,
  radio5_disabled: string,
  wifi_onoff: string,
  wifi_start_mode: string,
}
const wifiStatus = ref<WifiStatus>({} as WifiStatus);

interface SysVersion {
  ".anonymous": boolean,
  ".type": string,
  ".name": string,
  manufacturer: string,
  hardware_version: string,
  wa_inner_version: string,
  model_name: string,
  integrate_version: string,
  device_alias_name: string,
  imei_sv: string,
  device_market_name: string,
}
const sysVersion = ref<SysVersion>({} as SysVersion);

// USB状态
interface USBStatus {
  connect: number,
  mode: string,
  typec_cc: string,
  usb2rj45: number,
}
const usbStatus = ref<USBStatus>({} as USBStatus);

// 连接状态
interface NetAmbr {
  raw: string;
  dl: {
    value:    number;
    unit:     string;
    unit_num: number;
    unit_raw: string;
  };
  ul: {
    value:    number;
    unit:     string;
    unit_num: number;
    unit_raw: string;
  };
  qci1: number;
  qci2: number;
  iccid: string;
  sim_number: string;
}
const netAmbr = ref<NetAmbr>({
  raw: '',
  dl: {
    value: 0,
    unit: '',
    unit_num: 0,
    unit_raw: ''
  },
  ul: {
    value: 0,
    unit: '',
    unit_num: 0,
    unit_raw: ''
  },
  qci1: 0,
  qci2: 0,
  iccid: '',
  sim_number: ''
});



// 响应式数据
const loading = ref(false);
const error = ref<string | null>(null);
const data = ref<NetInfoResult | null>(null);
const lanData = ref<NetworkInterface>({} as NetworkInterface);
const wanData = ref<NetworkInterface>({} as NetworkInterface);
const wan6Data = ref<NetworkInterface>({} as NetworkInterface);
const trafficData = ref<TrafficData>({} as TrafficData);
const connTrack = ref<ConnTrackData | null>(null);
// conntrack-tune 安装状态（按钮 disabled / 文案切换来源）
interface ConntrackStatus {
  installed: boolean
  rc_link_exists: boolean
  script_path: string
}
const ctStatus = ref<ConntrackStatus | null>(null)
const ctInstallLoading = ref(false)
const ctUninstallLoading = ref(false)
// RF 频段功率按钮 loading
const rfSetLoading = ref(false)
const rfDefaultLoading = ref(false)
// RF 频段功率：来自 /api/sys/rf-maxpower/status
// state: default(默认) / applied_target(已应用29dBm) / custom(自定义非29) / unknown(读取失败)
const rfStatus = ref<{
  applied: boolean
  state: 'default' | 'applied_target' | 'custom' | 'unknown'
  recorded_target: boolean
  consistent: boolean
  drift_msg: string
  entries: any[]
} | null>(null)
const rfStateText = computed(() => {
  switch (rfStatus.value?.state) {
    case 'applied_target': return '● 已应用 29dBm（所有频段）'
    case 'custom': return '● 已应用自定义功率（非 29dBm）'
    case 'unknown': return '○ 状态读取失败'
    default: return '○ 当前为默认值'
  }
})
const deviceInfo = ref<DeviceInfo>({} as DeviceInfo);
const cpuTemp = ref<CpuTemp>({} as CpuTemp);
const simInfo = ref<SimInfo>({} as SimInfo);
const simInfo2 = ref<SimInfo2>({} as SimInfo2);
const wwanInfo = ref<WwanInfo>({} as WwanInfo);
const lanUserList = ref<LanUserList>({} as LanUserList);
const networkType = computed(() => {
  const val = d.value?.network_type;
  if (
    val?.includes('NR') ||
    val?.includes('5G') ||
    val?.includes('SA') ||
    val?.includes('NSA') ||
    val?.includes('ENDC')
  )
    return '5G';
  if (val?.includes('4G') || val?.includes('LTE')) return '4G';
  if (val?.includes('HSPA')) return 'H+';
  if (val?.includes('3G')) return '3G';
  return '';
});
const currentActiveBands = computed(() => {
  const val = networkType.value != '5G' ? data.value?.lteca : data.value?.nrca;
  if (!val) return null;
  const list = (val as string).split(';').filter((el) => el);
  if (list.length == 0) return null;
  if (networkType.value != '5G' && list.length == 1) return null;

  const res = list
    .map((item) =>
      networkType.value != '5G'
        ? item[1]
          ? item.split(',')[1]
          : ''
        : item[3]
        ? 'N' + item.split(',')[3]
        : ''
    )
    .join(', ')
    .replace(/,/g, ',');
  return res;
});

function formatSpeedUnit(unit?: string) {
  if (!unit) return '';

  const u = unit.toLowerCase();

  if (u.includes('mbps')) return 'M';
  if (u.includes('kbps')) return 'K';
  if (u.includes('bps')) return 'B';

  return unit;
}

const is5GA = computed(() => {
  if (!currentActiveBands.value) return false;
  return currentActiveBands.value.split(',').length >= 2;
});

// conntrack 占用率配色：<60% 绿，60-85% 橙，>85% 红
const ctUsageClass = computed(() => {
  const p = connTrack.value?.usage_percent ?? 0;
  if (p > 85) return 'ct-fill-danger';
  if (p >= 60) return 'ct-fill-warn';
  return 'ct-fill-safe';
});

// 自动刷新控制
const autoRefresh = ref(true);
const refreshInterval = ref(1000);
const refreshInterval2 = ref(5000);
let refreshTimer: number | null = null;
let refreshTimer2: number | null = null;
let adbTimer: number | null = null;
let firewallTimer: number | null = null;

// 请求体定义
// const netInfoRequest = {
//   id: 1,
//   service: 'zte_nwinfo_api',
//   method: 'nwinfo_get_netinfo',
//   params: {},
// };

// const lanRequest = {
//   id: 2,
//   service: 'network.interface.lan',
//   method: 'status',
//   params: {},
// };
//
// const wanRequest = {
//   id: 3,
//   service: 'network.interface.zte_wan',
//   method: 'status',
//   params: {},
// };
//
// const wan6Request = {
//   id: 4,
//   service: 'network.interface.zte_wan6',
//   method: 'status',
//   params: {},
// };
//
// const trafficRequest = {
//   id: 5,
//   service: 'zwrt_data',
//   method: 'get_wwandst',
//   params: { source_module: 'web', cid: 1, type: 4 },
// };
//
// const simInfoRequest = {
//   id: 6,
//   service: 'uci',
//   method: 'get',
//   params: {
//     config: 'zwrt_zte_mdm',
//     section: 'device_info',
//   },
// };
//
// const simInfo2Request = {
//   id: 7,
//   service: 'zwrt_zte_mdm.api',
//   method: 'get_sim_info',
//   params: {},
// };
//
// const deviceInfoRequest = {
//   id: 8,
//   service: 'zwrt_mc.device.manager',
//   method: 'get_device_info',
//   params: {},
// };
//
// const cpuTempRequest = {
//   id: 9,
//   service: 'zwrt_bsp.thermal',
//   method: 'get_cpu_temp',
//   params: {},
// };
//
// const wwanRequest = {
//   id: 10,
//   service: 'zwrt_data',
//   method: 'get_wwaniface',
//   params: {
//     source_module: 'web',
//     cid: 1,
//     connect_status: '',
//   },
// };
//
// const lanUserListRequest = {
//   id: 11,
//   service: 'zwrt_router.api',
//   method: 'router_get_user_list_num',
//   params: {},
// };
//
// const openAdbRequest = {
//   id: 12,
//   service: 'zwrt_bsp.usb',
//   method: 'set',
//   params: {
//     mode: 'debug',
//   },
// };
//
// const closeAdbRequest = {
//   id: 13,
//   service: 'zwrt_bsp.usb',
//   method: 'set',
//   params: {
//     mode: 'user',
//   },
// };

// session 固定值（未登录）
const SESSION_ID = '00000000000000000000000000000000'

// 1.网络信息
const netInfoRequest = {
  jsonrpc: '2.0',
  id: 1,
  method: 'call',
  params: [
    SESSION_ID,
    'zte_nwinfo_api',
    'nwinfo_get_netinfo',
    {},
  ]
};
// 2.LAN 状态
const lanRequest = {
  jsonrpc: '2.0',
  id: 2,
  method: 'call',
  params: [
    SESSION_ID,
    'network.interface.lan',
    'status',
    {},
  ],
}
// 3.WAN IPv4
const wanRequest = {
  jsonrpc: '2.0',
  id: 3,
  method: 'call',
  params: [
    SESSION_ID,
    'network.interface.zte_wan',
    'status',
    {},
  ],
}
// 4.WAN IPv6
const wan6Request = {
  jsonrpc: '2.0',
  id: 4,
  method: 'call',
  params: [
    SESSION_ID,
    'network.interface.zte_wan6',
    'status',
    {},
  ],
}
// 5.流量统计
const trafficRequest = {
  jsonrpc: '2.0',
  id: 5,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_data',
    'get_wwandst',
    { source_module: 'web', cid: 1, type: 4 },
  ],
}
// 6.设备信息
const deviceInfoRequest = {
  jsonrpc: '2.0',
  id: 6,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_mc.device.manager',
    'get_device_info',
    {},
  ],
}
// 7.CPU 温度
const cpuTempRequest = {
  jsonrpc: '2.0',
  id: 7,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_bsp.thermal',
    'get_cpu_temp',
    {},
  ],
}
// 8.SIM 信息（uci）
const simInfoRequest = {
  jsonrpc: '2.0',
  id: 8,
  method: 'call',
  params: [
    SESSION_ID,
    'uci',
    'get',
    {
      config: 'zwrt_zte_mdm',
      section: 'device_info',
    },
  ],
}
// 9.SIM 信息 2
const simInfo2Request = {
  jsonrpc: '2.0',
  id: 9,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_zte_mdm.api',
    'get_sim_info',
    {},
  ],
}
// 10.WWAN 接口信息
const wwanRequest = {
  jsonrpc: '2.0',
  id: 10,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_data',
    'get_wwaniface',
    {
      source_module: 'web',
      cid: 1,
      connect_status: '',
    },
  ],
}
// 11.LAN 用户数
const lanUserListRequest = {
  jsonrpc: '2.0',
  id: 11,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_router.api',
    'router_get_user_list_num',
    {},
  ],
}

// wifi 状态
const wifiStatusRequest = {
  jsonrpc: '2.0',
  id: 14,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_wlan',
    'report',
    {},
  ],
}

// 打开 ADB
// const openAdbRequest = {
//   jsonrpc: '2.0',
//   id: 12,
//   method: 'call',
//   params: [
//     SESSION_ID,
//     'zwrt_bsp.usb',
//     'set',
//     {
//       mode: 'debug',
//     },
//   ],
// }

// 关闭 ADB
// const closeAdbRequest = {
//   jsonrpc: '2.0',
//   id: 13,
//   method: 'call',
//   params: [
//     SESSION_ID,
//     'zwrt_bsp.usb',
//     'set',
//     {
//       mode: 'user',
//     },
//   ],
// }

// 系统版本信息 (G5 Pro: zwrt_common_info 在 U60 是 uci 配置, G5 Pro 改为 zwrt_zte_mdm.api 的 get_zwrt_common_info 方法)
const sysVersionRequest = {
  jsonrpc: '2.0',
  id: 15,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_zte_mdm.api',
    'get_zwrt_common_info',
    {},
  ],
};

const usbStatusRequest = {
  jsonrpc: '2.0',
  id: 16,
  method: 'call',
  params: [
    SESSION_ID,
    'zwrt_bsp.usb',
    'list',
    {},
  ],
};

// 1.网络信息 => netInfoRequest
// 2.LAN 状态 => lanRequest
// 3.WAN IPv4 => wanRequest
// 4.WAN IPv6 => wan6Request
// 5.流量统计 => trafficRequest
// 6.设备信息 => deviceInfoRequest
// 7.CPU 温度 => cpuTempRequest
// 8.SIM 信息（uci） => simInfoRequest
// 9.SIM 信息 2 => simInfo2Request
// 10.WWAN 接口信息 => wwanRequest
// 11.LAN 用户数 => lanUserListRequest
const batchRequests = [
  netInfoRequest,
  lanRequest,
  wanRequest,
  wan6Request,
  trafficRequest,
  // deviceInfoRequest,
  cpuTempRequest,
  simInfoRequest,
  simInfo2Request,
  wifiStatusRequest,
  sysVersionRequest,
  usbStatusRequest,
  // wwanRequest,
  // lanUserListRequest,
]
const batchRequests2 = [
  deviceInfoRequest,
  wwanRequest,
  lanUserListRequest
]

// 计算属性
const dataReady = computed(() => !!data.value);
const d = computed(() => data.value || {});

const signalBars = computed(() => {
  const bars = Number(d.value.signalbar || 0);
  if (Number.isNaN(bars)) return 0;
  return Math.max(0, Math.min(5, bars));
});

// 优先取 5G NR 字段（非 0 才用），否则回退 LTE；0 / 空视为无效
function pickByNet(nr: number | string, lte: number | string): number {
  const a = Number(nr);
  if (Number.isFinite(a) && a !== 0) return a;
  const b = Number(lte);
  if (Number.isFinite(b) && b !== 0) return b;
  return 0;
}
// 综合评分 4 维激活值（优先 NR，回退 LTE）
const activeRsrp = computed(() => pickByNet(d.value.nr5g_rsrp, d.value.lte_rsrp));
const activeRsrq = computed(() => pickByNet(d.value.nr5g_rsrq, d.value.lte_rsrq));
const activeSinr = computed(() => pickByNet(d.value.nr5g_snr,  d.value.lte_snr));
const activeRssi = computed(() => pickByNet(d.value.nr5g_rssi, d.value.lte_rssi));

// 信号强度文字评级（与 RSRP 评级保持一致的 4 档：优/良/中/差）
const signalRating = computed(() => {
  const r = activeRssi.value;
  if (!r) return { text: '未知', grade: 'unknown' };
  if (r >= -65) return { text: '优', grade: 'excellent' };
  if (r >= -75) return { text: '良', grade: 'good' };
  if (r >= -85) return { text: '中', grade: 'fair' };
  return { text: '差', grade: 'poor' };
});

const signalRatingText = computed(() => signalRating.value.text);

// AMBR 信号 tile 主题色：按强度等级切换配色（紫蓝→绿）
const signalTileClass = computed(() => `grade-${signalRating.value.grade}`);

// ------------------------------------------------------------------
// 网络综合评分（业界主流 LTE/5G 评估法）
// 维度：RSRP(覆盖强度) / RSRQ(信号质量) / SINR(信噪比) / RSSI(接收强度)
// 各维度按 3GPP 阈值线性映射到 0-100，再加权求和
// 权重参考网络优化软件常见配置：RSRP 主导覆盖 → 40%，SINR 决定吞吐 → 30%，
//                              RSRQ 质量补充 → 20%，RSSI 辅助(含噪声) → 10%
// ------------------------------------------------------------------
// RSRP (dBm)：≥-80 优, -90 良, -100 中, -110 差, <-120 极差
function scoreRsrp(v: number): number {
  if (!v) return 0;
  if (v >= -80) return 100;
  if (v >= -90) return 80 + (v + 90) * 2;
  if (v >= -100) return 60 + (v + 100) * 2;
  if (v >= -110) return 40 + (v + 110) * 2;
  if (v >= -120) return 20 + (v + 120) * 2;
  return Math.max(0, 20 + (v + 120) * 0.5);
}
// RSRQ (dB)：≥-10 优, -15 良, -20 中, <-20 差
function scoreRsrq(v: number): number {
  if (!v) return 0;
  if (v >= -10) return 100;
  if (v >= -15) return 70 + (v + 15) * 6;
  if (v >= -20) return 30 + (v + 20) * 8;
  return Math.max(0, 30 + (v + 20) * 1.5);
}
// SINR (dB)：≥20 优, 13 良, 0 中, -10 差, <-10 极差
function scoreSinr(v: number): number {
  if (!v) return 0;
  if (v >= 20) return 100;
  if (v >= 13) return 80 + (v - 13) * (20 / 7);
  if (v >= 0)  return 40 + v * (40 / 13);
  if (v >= -10) return 10 + (v + 10) * 3;
  return Math.max(0, 10 + (v + 10) * 0.5);
}
// RSSI (dBm)：≥-65 优, -85 良, -100 中, -120 差
function scoreRssi(v: number): number {
  if (!v) return 0;
  if (v >= -65) return 100;
  if (v >= -85) return 70 + (v + 85) * 1.5;
  if (v >= -100) return 40 + (v + 100) * 2;
  if (v >= -120) return 10 + (v + 120) * 1.5;
  return Math.max(0, 10 + (v + 120) * 0.5);
}

const networkScore = computed(() => {
  const a = scoreRsrp(activeRsrp.value);   // 40%
  const b = scoreSinr(activeSinr.value);   // 30%
  const c = scoreRsrq(activeRsrq.value);   // 20%
  const e = scoreRssi(activeRssi.value);   // 10%
  return Math.round(a * 0.4 + b * 0.3 + c * 0.2 + e * 0.1);
});

// 综合评分评级（4 档；阈值按主流综合评分分级：≥85 优 / ≥70 良 / ≥55 中 / 否则差）
const networkScoreRating = computed(() => {
  const s = networkScore.value;
  if (!s && s !== 0) return { text: '未知', grade: 'unknown' };
  if (s >= 85) return { text: '优', grade: 'excellent' };
  if (s >= 70) return { text: '良', grade: 'good' };
  if (s >= 55) return { text: '中', grade: 'fair' };
  return { text: '差', grade: 'poor' };
});
const networkScoreText      = computed(() => networkScoreRating.value.text);
const networkScoreTileClass = computed(() => `grade-${networkScoreRating.value.grade}`);
// 综合评分字体色（按档位变化）
const networkScoreColor = computed(() => {
  const g = networkScoreRating.value.grade;
  if (g === 'excellent') return '#7dffb0';
  if (g === 'good')      return '#8be9ff';
  if (g === 'fair')      return '#ffb84d';
  if (g === 'poor')      return '#ff7a7a';
  return '#cbd5e0';
});

// QCI/5QI 业务等级中文提示（让 QCI tile 不太空，与其它 tile 信息密度对齐）
// 主要给出 GBR/非 GBR 类别标签；具体业务类型以运营商下发的 APN 配置为准
const QCI_HINT: Record<number, string> = {
  1: '语音会话(GBR)',
  2: '视频会话(GBR)',
  3: '游戏交互(GBR)',
  4: '视频流(GBR)',
  5: 'IMS信令(GBR)',
  6: '视频·低时延',
  7: '语音视频',
  8: '视频缓冲',
  9: '默认数据',
};
const qciHint = computed(() => {
  const v = netAmbr.value?.qci2 || netAmbr.value?.qci1;
  const n = Number(v);
  if (!Number.isFinite(n) || n <= 0) return '业务等级';
  if (n === 9) return ''; // 默认承载（默认数据），副标题无意义，隐藏
  return QCI_HINT[n] || `5QI ${n}`;
});


// 格式化函数
function formatDbm(v: unknown): string {
  const n = Number(v);
  if (Number.isNaN(n)) return '-';
  return `${n} dBm`;
}

function formatDb(v: unknown): string {
  const n = Number(v);
  if (Number.isNaN(n)) return '-';
  return `${n} dB`;
}

function formatSnr(v: unknown): string {
  const n = Number(v);
  if (Number.isNaN(n)) return '-';
  return `${n.toFixed(1)} dB`;
}

// 仅显示数值（如 "15.0"），不带 dB 单位，用于副指标行紧凑展示
function formatSnrNoUnit(v: unknown): string {
  const n = Number(v);
  if (Number.isNaN(n)) return '-';
  return n.toFixed(1);
}

// 综合评分副指标用：仅显示整数测量值（RSRP/RSRQ/RSSI 单位为 dBm/dB，由 label 区分）
function fmtInt(v: unknown): string {
  const n = Number(v);
  if (!Number.isFinite(n) || n === 0) return '-';
  return `${n}`;
}

function formatUptime(seconds?: number): string {
  if (!seconds || seconds <= 0) return '-';

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  const parts: string[] = [];
  if (days) parts.push(`${days}天`);
  parts.push(`${hours}小时`);
  parts.push(`${minutes}分`);
  // if (secs || parts.length === 0) parts.push(`${secs}秒`);
  return parts.join('');
}

// 速率统一以比特(bit)为单位显示, 与顶部 AMBR 的 Mbps 保持一致
function formatSpeed(bytesPerSecond: number): string {
  if (!bytesPerSecond) return '0 b/s';
  const units = ['b/s', 'Kb/s', 'Mb/s', 'Gb/s'];
  let size = bytesPerSecond * 8; // 字节 -> 比特
  let unitIndex = 0;

  while (size >= 1000 && unitIndex < units.length - 1) {
    size /= 1000;
    unitIndex++;
  }

  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

function formatBytes(bytes: number): string {
  if (!bytes) return '-';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

function formatNumber(num: number): string {
  if (!num) return '-';
  return num.toLocaleString();
}

// 系统信息格式化函数
function formatLoad(load: number): string {
  return (load / 1000).toFixed(2);
}

function formatMemory(KB: number): string {
  if (!KB) return '-';
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = KB * 1024;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

function formatMemoryPercent(used: number, total: number): number {
  if (!total) return 0;
  return Math.round((used / total) * 100);
}

function formatCpuTemp(temp: number): string {
  if (!temp) return '-';
  return `${temp}°C`;
}

// 信号强度百分比计算函数
function getRssiPercent(rssi: number): number {
  if (!rssi) return 0;
  // RSSI: -120dBm (0%) 到 -30dBm (100%)
  const min = -120;
  const max = -30;
  const percent = Math.max(
    0,
    Math.min(100, ((rssi - min) / (max - min)) * 100)
  );
  return Math.round(percent);
}

function getRsrpPercent(rsrp: number): number {
  if (!rsrp) return 0;
  // RSRP: -140dBm (0%) 到 -70dBm (100%)
  const min = -140;
  const max = -70;
  const percent = Math.max(
    0,
    Math.min(100, ((rsrp - min) / (max - min)) * 100)
  );
  return Math.round(percent);
}

function getRsrqPercent(rsrq: number): number {
  if (!rsrq) return 0;
  // RSRQ: -20dB (0%) 到 -3dB (100%)
  const min = -20;
  const max = -3;
  const percent = Math.max(
    0,
    Math.min(100, ((rsrq - min) / (max - min)) * 100)
  );
  return Math.round(percent);
}

function getSnrPercent(snr: number): number {
  if (!snr) return 0;
  // SNR: 0dB (0%) 到 30dB (100%)
  const min = 0;
  const max = 30;
  const percent = Math.max(0, Math.min(100, ((snr - min) / (max - min)) * 100));
  return Math.round(percent);
}

type SignalMetric = 'rsrp' | 'rsrq' | 'sinr' | 'rssi';
type SignalGrade = 'excellent' | 'good' | 'fair' | 'poor' | 'unknown';
type SignalType = 'nr' | 'lte';
type SignalHelpKey = `${SignalType}-${SignalMetric}`;

interface SignalStatus {
  text: string;
  className: SignalGrade;
  score: number;
}

interface SignalHelpRange {
  label: string;
  className: SignalGrade;
}

interface SignalHelp {
  title: string;
  description: string;
  ranges: SignalHelpRange[];
}

const signalStatusMap: Record<SignalGrade, SignalStatus> = {
  excellent: { text: '优秀', className: 'excellent', score: 4 },
  good: { text: '良好', className: 'good', score: 3 },
  fair: { text: '一般', className: 'fair', score: 2 },
  poor: { text: '较差', className: 'poor', score: 1 },
  unknown: { text: '未知', className: 'unknown', score: 0 },
};

const signalHelpMap: Record<SignalMetric, SignalHelp> = {
  rsrp: {
    title: 'RSRP：信号覆盖强度',
    description: '主要看基站信号到设备这里有多强。数值是负数，越接近 0 越好。',
    ranges: [
      { label: '≥ -85 dBm：优秀', className: 'excellent' },
      { label: '-95 到 -86 dBm：良好', className: 'good' },
      { label: '-105 到 -96 dBm：一般', className: 'fair' },
      { label: '< -105 dBm：较差', className: 'poor' },
    ],
  },
  rsrq: {
    title: 'RSRQ：信号质量',
    description: '主要看信号是否干净、是否拥挤。数值也是负数，越接近 0 越好。',
    ranges: [
      { label: '≥ -8 dB：优秀', className: 'excellent' },
      { label: '-11 到 -9 dB：良好', className: 'good' },
      { label: '-15 到 -12 dB：一般', className: 'fair' },
      { label: '< -15 dB：较差', className: 'poor' },
    ],
  },
  sinr: {
    title: 'SINR：信噪比',
    description: '主要看有用信号比干扰和噪声强多少。数值越大越好。',
    ranges: [
      { label: '≥ 20 dB：优秀', className: 'excellent' },
      { label: '13 到 19.9 dB：良好', className: 'good' },
      { label: '> 0 到 12.9 dB：一般', className: 'fair' },
      { label: '≤ 0 dB：较差', className: 'poor' },
    ],
  },
  rssi: {
    title: 'RSSI：接收总强度',
    description: '包含有用信号、干扰和噪声，只适合作辅助参考。数值越接近 0 越好。',
    ranges: [
      { label: '≥ -65 dBm：优秀', className: 'excellent' },
      { label: '-75 到 -66 dBm：良好', className: 'good' },
      { label: '-85 到 -76 dBm：一般', className: 'fair' },
      { label: '< -85 dBm：较差', className: 'poor' },
    ],
  },
};

const openedSignalHelp = ref<SignalHelpKey | null>(null);

function getSignalHelpKey(type: SignalType, metric: SignalMetric): SignalHelpKey {
  return `${type}-${metric}`;
}

function isSignalHelpOpen(type: SignalType, metric: SignalMetric): boolean {
  return openedSignalHelp.value === getSignalHelpKey(type, metric);
}

function toggleSignalHelp(type: SignalType, metric: SignalMetric) {
  const key = getSignalHelpKey(type, metric);
  openedSignalHelp.value = openedSignalHelp.value === key ? null : key;
}

function getSignalStatus(metric: SignalMetric, rawValue: unknown): SignalStatus {
  if (rawValue === null || rawValue === undefined || rawValue === '') {
    return signalStatusMap.unknown;
  }
  const value = Number(rawValue);
  if (!Number.isFinite(value)) return signalStatusMap.unknown;

  if (metric === 'rsrp') {
    if (value === 0) return signalStatusMap.unknown;
    if (value >= -85) return signalStatusMap.excellent;
    if (value >= -95) return signalStatusMap.good;
    if (value >= -105) return signalStatusMap.fair;
    return signalStatusMap.poor;
  }

  if (metric === 'rsrq') {
    if (value === 0) return signalStatusMap.unknown;
    if (value >= -8) return signalStatusMap.excellent;
    if (value >= -11) return signalStatusMap.good;
    if (value >= -15) return signalStatusMap.fair;
    return signalStatusMap.poor;
  }

  if (metric === 'sinr') {
    if (value >= 20) return signalStatusMap.excellent;
    if (value >= 13) return signalStatusMap.good;
    if (value > 0) return signalStatusMap.fair;
    return signalStatusMap.poor;
  }

  if (value === 0) return signalStatusMap.unknown;
  if (value >= -65) return signalStatusMap.excellent;
  if (value >= -75) return signalStatusMap.good;
  if (value >= -85) return signalStatusMap.fair;
  return signalStatusMap.poor;
}

function hasUsableSignalValue(rawValue: unknown): boolean {
  if (rawValue === null || rawValue === undefined || rawValue === '') {
    return false;
  }
  const value = Number(rawValue);
  return Number.isFinite(value) && value !== 0;
}

function isSignalActive(type: 'nr' | 'lte'): boolean {
  if (type === 'nr') {
    return networkType.value === '5G' && hasUsableSignalValue(d.value.nr5g_rsrp);
  }
  return networkType.value === '4G' && hasUsableSignalValue(d.value.lte_rsrp);
}

function getSignalDisplayStatus(
  type: 'nr' | 'lte',
  metric: SignalMetric,
  rawValue: unknown
): SignalStatus {
  if (!isSignalActive(type)) return signalStatusMap.unknown;
  return getSignalStatus(metric, rawValue);
}

function getAverageSignalStatus(statuses: SignalStatus[]): SignalStatus {
  const validStatuses = statuses.filter(item => item.className !== 'unknown');
  if (!validStatuses.length) return signalStatusMap.unknown;

  const averageScore =
    validStatuses.reduce((total, item) => total + item.score, 0) /
    validStatuses.length;

  let status: SignalStatus;
  if (averageScore >= 3.5) {
    status = signalStatusMap.excellent;
  } else if (averageScore >= 2.5) {
    status = signalStatusMap.good;
  } else if (averageScore >= 1.5) {
    status = signalStatusMap.fair;
  } else {
    status = signalStatusMap.poor;
  }

  const weakestScore = Math.min(...validStatuses.map(item => item.score));
  if (weakestScore <= 1 && status.score > signalStatusMap.fair.score) {
    return signalStatusMap.fair;
  }
  if (weakestScore <= 2 && status.score > signalStatusMap.good.score) {
    return signalStatusMap.good;
  }
  return status;
}

function getNetworkSignalStatus(type: 'nr' | 'lte'): SignalStatus {
  if (!isSignalActive(type)) return signalStatusMap.unknown;

  if (type === 'nr') {
    return getAverageSignalStatus([
      getSignalStatus('rsrp', d.value.nr5g_rsrp),
      getSignalStatus('rsrq', d.value.nr5g_rsrq),
      getSignalStatus('sinr', d.value.nr5g_snr),
      getSignalStatus('rssi', d.value.nr5g_rssi),
    ]);
  }

  return getAverageSignalStatus([
    getSignalStatus('rsrp', d.value.lte_rsrp),
    getSignalStatus('rsrq', d.value.lte_rsrq),
    getSignalStatus('sinr', d.value.lte_snr),
    getSignalStatus('rssi', d.value.lte_rssi),
  ]);
}

// 获取载波信息
function formatNrca(nrca: string, pre: string, type: number, index: number): string {
  if (!nrca) return '-';
  const carriers = nrca.split(';').filter(item => item.trim() !== '');
  const carrier = carriers[type];
  if (!carrier) return '-';
  const fields = carrier.split(',');
  // index 越界
  if (index < 0 || index >= fields.length) return '-';
  const value = fields[index];
  // 参数为空
  if (!value || value.trim() === '') return '-';
  const num = Number(value);
  if (Number.isNaN(num)) return '-';
  return pre + String(num);
}

// 获取网络运营商
const netWorkProvider = computed(() => {
  let provider = data.value?.network_provider;
  const fullname = data.value?.network_provider_fullname;
  const Operator = simInfo2?.value?.Operator;

  if (!provider && !fullname) return '-';
  // 中国联通 特殊处理
  if (provider === 'UNICOM') provider = 'CUCC';
  const providerMap: Record<string, string> = {
    CMCC: '中国移动',
    CUCC: '中国联通',
    UNICOM: '中国联通',
    CT: '中国电信',
    CBN: '中国广电',
  };
  // 优先走 code 映射，映射不到再走后端全名
  return providerMap[provider] ?? fullname ?? '-';
  // return (providerMap[Operator] ?? Operator ) + (Operator === provider ? '' : '(' + (providerMap[provider] ?? fullname ?? '') + ')');
});

// API 调用函数
// async function callUbus<T>(request: any): Promise<T> {
//   const response = await axios.post<UbusResponse<T>>('/api/ubus', request);
//   if (response.data.code === 0) {
//     return response.data.result as T;
//   } else {
//     throw new Error(response.data.msg || '接口返回错误');
//   }
// }

async function callUbusBatch(
    requests: any[]
): Promise<Record<number, any>> {
  const resp = await axios.post<ZteRpcResponse[]>(
      '/api/ubus', requests
  )
  const map: Record<number, any> = {}
  for (const item of resp.data) {
    const [code, data] = item.result
    if (code === 0) {
      map[item.id] = data
    } else {
      console.error(`ubus call failed, id=${item.id}`, data)
    }
  }
  return map
}

// async function fetchAllData() {
//   loading.value = true;
//   error.value = null;
//
//   try {
//     // 并行请求所有数据
//     const [
//       netInfo,
//       lan,
//       wan,
//       wan6,
//       traffic,
//       device,
//       cpuTempData,
//       simInfoData,
//       simInfo2Data,
//       wwanInfoData,
//       lanUserData,
//     ] = await Promise.all([
//       callUbus<NetInfoResult>(netInfoRequest),
//       callUbus<NetworkInterface>(lanRequest),
//       callUbus<NetworkInterface>(wanRequest),
//       callUbus<NetworkInterface>(wan6Request),
//       callUbus<TrafficData>(trafficRequest),
//       callUbus<DeviceInfo>(deviceInfoRequest),
//       callUbus<CpuTemp>(cpuTempRequest),
//       callUbus<SimInfo>(simInfoRequest),
//       callUbus<SimInfo2>(simInfo2Request),
//       callUbus<WwanInfo>(wwanRequest),
//       callUbus<LanUserList>(lanUserListRequest),
//     ]);
//
//     data.value = netInfo;
//     lanData.value = lan;
//     wanData.value = wan;
//     wan6Data.value = wan6;
//     trafficData.value = traffic;
//     deviceInfo.value = device;
//     cpuTemp.value = cpuTempData;
//     simInfo.value = simInfoData;
//     simInfo2.value = simInfo2Data;
//     wwanInfo.value = wwanInfoData;
//     lanUserList.value = lanUserData;
//   } catch (e: any) {
//     error.value = e?.message || '请求失败';
//     console.error('数据获取失败:', e);
//   } finally {
//     loading.value = false;
//   }
// }

async function fetchAllData() {
  loading.value = true
  error.value = null
  try {
    const resultMap = await callUbusBatch(batchRequests)
    // 按 id 取值（清晰又稳定）
    data.value        = resultMap[1]
    lanData.value     = resultMap[2]
    wanData.value     = resultMap[3]
    wan6Data.value    = resultMap[4]
    trafficData.value = resultMap[5]
    // deviceInfo.value  = resultMap[6]
    cpuTemp.value     = resultMap[7]
    simInfo.value     = resultMap[8]
    simInfo2.value    = resultMap[9]
    // wwanInfo.value    = resultMap[10]
    // lanUserList.value = resultMap[11]
    wifiStatus.value = resultMap[14]
    // G5 Pro: get_zwrt_common_info 直接返回扁平对象(非 uci 的 {values:{...}} 结构)
    sysVersion.value = resultMap[15] ?? {}
    usbStatus.value = resultMap[16]
  } catch (e: any) {
    error.value = e?.message || '请求失败'
    console.error('数据获取失败:', e)
    data.value = null
    lanData.value = {} as NetworkInterface
    wanData.value = {} as NetworkInterface
    wan6Data.value = {} as NetworkInterface
    trafficData.value = {} as TrafficData
    cpuTemp.value = {} as CpuTemp
    simInfo.value = {} as SimInfo
    simInfo2.value = {} as SimInfo2
    wifiStatus.value = {} as WifiStatus
    sysVersion.value = {} as SysVersion
    usbStatus.value = {} as USBStatus
  } finally {
    loading.value = false
  }
}
async function fetchAllData2() {
  loading.value = true
  error.value = null
  try {
    const [resultMap, ct] = await Promise.all([
      callUbusBatch(batchRequests2),
      axios
        .get<{ code: number; data: ConnTrackData }>('/api/sys/conntrack')
        .then((r) => (r.data && r.data.code === 0 ? r.data.data : null))
        .catch(() => null),
    ])
    // 按 id 取值（清晰又稳定）
    // data.value        = resultMap[1]
    // lanData.value     = resultMap[2]
    // wanData.value     = resultMap[3]
    // wan6Data.value    = resultMap[4]
    // trafficData.value = resultMap[5]
    deviceInfo.value  = resultMap[6]
    // cpuTemp.value     = resultMap[7]
    // simInfo.value     = resultMap[8]
    // simInfo2.value    = resultMap[9]
    wwanInfo.value    = resultMap[10]
    lanUserList.value = resultMap[11]
    if (ct) connTrack.value = ct
    // 顺带刷新 conntrack-tune 安装状态（用于装机/卸载按钮的禁用与文案）
    axios
      .get<{ code: number } & ConntrackStatus>('/api/sys/conntrack/status')
      .then((r) => {
        if (r.data && (r.data as any).code === 0) {
          ctStatus.value = {
            installed: !!(r.data as any).installed,
            rc_link_exists: !!(r.data as any).rc_link_exists,
            script_path: (r.data as any).script_path,
          }
        }
      })
      .catch(() => { /* 取不到也不阻断主刷新 */ })
    // 顺带刷新 RF 频段功率状态（用于两个功率按钮的禁用与文案）
    axios
      .get<{ code: number; applied: boolean; entries: any[] }>('/api/sys/rf-maxpower/status')
      .then((r) => {
        if (r.data && r.data.code === 0) {
          rfStatus.value = {
            applied: !!r.data.applied,
            state: (r.data as any).state || 'unknown',
            recorded_target: !!(r.data as any).recorded_target,
            consistent: (r.data as any).consistent !== false,
            drift_msg: (r.data as any).drift_msg || '',
            entries: (r.data as any).entries || [],
          }
        }
      })
      .catch(() => { /* 取不到也不阻断主刷新 */ })
  } catch (e: any) {
    error.value = e?.message || '请求失败'
    console.error('数据获取失败:', e)
    deviceInfo.value = {} as DeviceInfo
    wwanInfo.value = {} as WwanInfo
    lanUserList.value = {} as LanUserList
  } finally {
    loading.value = false
  }

}

// conntrack-tune 装机：写 /etc/init.d/conntrack-tune + enable + start
async function onInstallConntrack() {
  ctInstallLoading.value = true
  try {
    const res = await axios.post<{
      code: number
      success: boolean
      msg: string
      script_path?: string
      rc_link?: string
      chmod?: number
      enable?: number
      start?: number
      enable_out?: string
      start_out?: string
      applied_at?: string
      steps?: Array<Record<string, any>>
    }>('/api/sys/conntrack/install')
    const ok = !!(res.data?.success)
    openActionDialog(buildConntrackInstallDialog(res.data, ok))
    if (ok) {
      ElMessage.success('已安装 conntrack-tune · 查看详情')
      fetchAllData2()
    } else {
      ElMessage.error(res.data?.msg || '装机失败')
    }
  } catch (e: any) {
    openActionDialog(buildErrorDialog('安装 conntrack-tune 失败', e))
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
  } finally {
    ctInstallLoading.value = false
    // 不论成败都查一次 status，让按钮文案/disabled 同步
    axios.get('/api/sys/conntrack/status').then((r) => {
      if (r.data && r.data.code === 0) {
        ctStatus.value = {
          installed: !!r.data.installed,
          rc_link_exists: !!r.data.rc_link_exists,
          script_path: r.data.script_path,
        }
      }
    }).catch(() => {})
  }
}

// conntrack-tune 卸载：disable + stop + rm
async function onUninstallConntrack() {
  try {
    await ElMessageBox.confirm(
      '将禁用并删除 conntrack-tune 脚本（重启后失效，恢复需重装）。\n旧名 zz-conntrack 残留也会一并清理。是否继续？',
      '卸载确认',
      { confirmButtonText: '卸载', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return // 用户取消
  }
  ctUninstallLoading.value = true
  try {
    const res = await axios.post<{ code: number; success: boolean; msg: string; steps?: any[] }>(
      '/api/sys/conntrack/uninstall',
    )
    const ok = !!(res.data?.success)
    openActionDialog(buildConntrackUninstallDialog(res.data, ok))
    if (ok) {
      ElMessage.success('已卸载 conntrack-tune · 查看详情')
      fetchAllData2()
    } else {
      ElMessage.error(res.data?.msg || '卸载失败')
    }
  } catch (e: any) {
    openActionDialog(buildErrorDialog('卸载 conntrack-tune 失败', e))
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
  } finally {
    ctUninstallLoading.value = false
    axios.get('/api/sys/conntrack/status').then((r) => {
      if (r.data && r.data.code === 0) {
        ctStatus.value = {
          installed: !!r.data.installed,
          rc_link_exists: !!r.data.rc_link_exists,
          script_path: r.data.script_path,
        }
      }
    }).catch(() => {})
  }
}

// RF 频段功率：所有频段设置为 29dBm（mipc_wan_cli --at_cmd AT+EFMAXPWR=...）
async function onSetRfMaxpower() {
  rfSetLoading.value = true
  try {
    const res = await axios.post<{ code: number; success: boolean; msg: string; details?: any[] }>(
      '/api/sys/rf-maxpower/set',
    )
    const ok = !!(res.data?.success)
    openActionDialog(buildRfDialog('所有频段功率设为 29dBm', res.data, ok))
    if (ok) ElMessage.success(res.data?.msg || '所有频段功率已设置为 29dBm · 查看详情')
    else ElMessage.warning(res.data?.msg || '部分命令失败，请查看详情')
  } catch (e: any) {
    openActionDialog(buildErrorDialog('设置 RF 频段功率失败', e))
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
  } finally {
    rfSetLoading.value = false
    fetchRfStatus()
  }
}

// RF 频段功率：恢复默认（AT 末位 0）
async function onSetRfDefault() {
  rfDefaultLoading.value = true
  try {
    const res = await axios.post<{ code: number; success: boolean; msg: string; details?: any[] }>(
      '/api/sys/rf-maxpower/default',
    )
    const ok = !!(res.data?.success)
    openActionDialog(buildRfDialog('恢复 RF 频段默认功率', res.data, ok))
    if (ok) ElMessage.success(res.data?.msg || '所有频段功率已恢复默认 · 查看详情')
    else ElMessage.warning(res.data?.msg || '部分命令失败，请查看详情')
  } catch (e: any) {
    openActionDialog(buildErrorDialog('恢复 RF 频段默认失败', e))
    ElMessage.error('请求失败：' + (e?.message || '未知错误'))
  } finally {
    rfDefaultLoading.value = false
    fetchRfStatus()
  }
}

// 系统操作结果弹窗：统一的状态/明细展示
interface ActionResultItem {
  label: string
  value?: string
  code?: number
  status?: 'ok' | 'fail' | 'warn'
}
interface ActionResultSection {
  title?: string
  items: ActionResultItem[]
}
interface ActionResultState {
  title: string
  success: boolean
  summary: string
  sections: ActionResultSection[]
  rawDetails?: string[]
}
const actionDialogVisible = ref(false)
const actionDialog = ref<ActionResultState>({
  title: '',
  success: true,
  summary: '',
  sections: [],
})
function openActionDialog(state: ActionResultState) {
  actionDialog.value = state
  actionDialogVisible.value = true
}
function copyActionRaw() {
  const text = (actionDialog.value.rawDetails || []).join('\n')
  if (!text) return
  // 优先用 navigator.clipdown，无可用文本时不阻塞
  if (navigator.clipboard?.writeText) {
    navigator.clipboard.writeText(text).catch(() => {})
  } else {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    try { document.execCommand('copy') } catch {}
    ta.remove()
  }
  ElMessage.success('已复制到剪贴板')
}

// 通用条目规范化（steps/details 数组 → UI 行）
function normalizeSteps(steps?: any[]): ActionResultSection {
  const items: ActionResultItem[] = []
  if (!steps) {
    return { title: '步骤', items }
  }
  for (const s of steps) {
    const status: ActionResultItem['status'] = s.ok
      ? 'ok'
      : s.err
        ? 'fail'
        : undefined
    let val = s.out
    if (val === undefined || val === '') {
      val = s.ok ? 'OK' : s.err || '—'
    }
    items.push({
      label: s.step || '?',
      value: val,
      code: s.code,
      status,
    })
  }
  return { title: '执行步骤', items }
}
function buildConntrackInstallDialog(d: any, ok: boolean): ActionResultState {
  const sections: ActionResultSection[] = []
  // 老名清理作为独立分区展示
  if (d?.steps?.[0]?.legacy_cleanup?.length) {
    sections.push(normalizeSteps(d.steps[0].legacy_cleanup))
  }
  if (d?.steps?.length > 1) {
    sections.push(normalizeSteps(d.steps.slice(1)))
  } else if (d?.steps && !sections.length) {
    sections.push(normalizeSteps(d.steps))
  }
  const metaItems: ActionResultItem[] = []
  if (d?.script_path) metaItems.push({ label: '脚本路径', value: d.script_path })
  if (d?.rc_link) metaItems.push({ label: '开机自启软链', value: d.rc_link })
  if (d?.applied_at) metaItems.push({ label: '应用时间', value: d.applied_at })
  if (d?.enable !== undefined) metaItems.push({ label: 'enable 返回码', value: String(d.enable) })
  if (d?.start !== undefined) metaItems.push({ label: 'start 返回码', value: String(d.start) })
  if (d?.enable_out) metaItems.push({ label: 'enable 输出', value: d.enable_out })
  if (d?.start_out) metaItems.push({ label: 'start 输出', value: d.start_out })
  if (metaItems.length) sections.push({ title: '元数据', items: metaItems })
  return {
    title: ok ? '安装 conntrack-tune' : '安装失败',
    success: ok,
    summary: ok ? '已写入 /etc/init.d/conntrack-tune 并 enable + start，开机自启 + 立即生效' : (d?.msg || '失败'),
    sections,
  }
}
function buildConntrackUninstallDialog(d: any, ok: boolean): ActionResultState {
  const sections: ActionResultSection[] = []
  if (d?.steps?.length) sections.push(normalizeSteps(d.steps))
  return {
    title: ok ? '卸载 conntrack-tune' : '卸载失败',
    success: ok,
    summary: ok ? '已 disable + stop + rm 脚本与软链，重启后系统参数会回退到出厂默认值' : (d?.msg || '失败'),
    sections,
  }
}
function buildRfDialog(title: string, d: any, ok: boolean): ActionResultState {
  const sections: ActionResultSection[] = []
  // details[]（AT 命令数组）：每条转成独立段落
  if (Array.isArray(d?.details) && d.details.length) {
    const items: ActionResultItem[] = []
    for (const dt of d.details) {
      items.push({
        label: dt.at,
        value: dt.out || (dt.err ? dt.err : '(no output)'),
        code: dt.code,
        status: dt.err || (dt.code !== undefined && dt.code !== 0) ? 'fail' : 'ok',
      })
    }
    sections.push({ title: 'AT 命令逐条执行结果', items })
  } else {
    sections.push({
      title: '执行结果',
      items: [{ label: '详情', value: '（无 details 返回）', status: 'warn' }],
    })
  }
  // 末尾加一段说明
  sections.push({
    title: '补充说明',
    items: [
      { label: '命令工具', value: 'mipc_wan_cli --at_cmd AT+EFMAXPWR=...' },
      { label: '生效位置', value: 'modem 寄存器（不依赖本工具维持，重启/网络重注册可能回到默认）' },
    ],
  })
  return {
    title,
    success: ok,
    summary: d?.msg || (ok ? '全部命令已成功' : '部分失败'),
    sections,
  }
}
function buildErrorDialog(title: string, e: any): ActionResultState {
  return {
    title,
    success: false,
    summary: e?.message || '网络请求失败',
    sections: [
      {
        title: '异常',
        items: [
          { label: '错误信息', value: e?.message || '未知', status: 'fail' },
          { label: '错误对象', value: String(e), status: 'fail' },
        ],
      },
    ],
  }
}


// RF 频段功率状态：读 modem EFMAXPWR，判断当前是否已应用自定义上限
async function fetchRfStatus() {
  try {
    const r = await axios.get<{ code: number; applied: boolean; entries: any[]; msg?: string }>(
      '/api/sys/rf-maxpower/status',
    )
    if (r.data && r.data.code === 0) {
      rfStatus.value = {
        applied: !!r.data.applied,
        state: (r.data as any).state || 'unknown',
        recorded_target: !!(r.data as any).recorded_target,
        consistent: (r.data as any).consistent !== false,
        drift_msg: (r.data as any).drift_msg || '',
        entries: (r.data as any).entries || [],
      }
    }
  } catch {
    /* 取不到也不阻断主刷新 */
  }
}

function refresh() {
  fetchAllData().then((res) => {
    ElMessage.success('数据已刷新');
  });
  fetchAllData2();
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value;

  if (autoRefresh.value) {
    startAutoRefresh();
    ElMessage.success('已恢复刷新');
  } else {
    stopAutoRefresh();
    ElMessage.warning('已停止刷新');
  }
}

function updateRefreshInterval() {
  if (autoRefresh.value) {
    stopAutoRefresh();
    startAutoRefresh();
  }
}

function startAutoRefresh() {
  stopAutoRefresh();
  refreshTimer = window.setInterval(() => {
    fetchAllData();
  }, refreshInterval.value);
  refreshTimer2 = window.setInterval(() => {
    fetchAllData2();
  }, refreshInterval2.value);
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
  if (refreshTimer2) {
    clearInterval(refreshTimer2);
    refreshTimer2 = null;
  }
}

const adbStatus = ref(false)
function refreshAdbStatus() {
  axios.get('/api/adb/status')
    .then(res => {
      if (res?.data?.code === 0) adbStatus.value = !!res.data.enabled
    })
    .catch(() => {})
}

// 防火墙开关状态：null=未知/读取失败；true=开；false=关
const firewallStatus = ref<boolean | null>(null)
function refreshFirewallStatus() {
  axios.get('/api/sys/firewall/status')
    .then(res => {
      if (res?.data?.code === 0 && res.data.data?.known) {
        firewallStatus.value = !!res.data.data.enabled
      }
    })
    .catch(() => {})
}

// 温度状态(G5 Pro 无电池, 直接读 sysfs thermal_zone, Modem 取代原电池栏)
const thermal = ref<{ cpu_temp: number; soc_max: number; modem_temp: number; board_temp: number }>({
  cpu_temp: 0, soc_max: 0, modem_temp: 0, board_temp: 0,
})
let thermalTimer: number | null = null
function refreshThermal() {
  axios.get('/api/sys/thermal')
    .then(res => {
      if (res?.data?.code === 0) {
        thermal.value = {
          cpu_temp: Number(res.data.cpu_temp) || 0,
          soc_max: Number(res.data.soc_max) || 0,
          modem_temp: Number(res.data.modem_temp) || 0,
          board_temp: Number(res.data.board_temp) || 0,
        }
      }
    })
    .catch(() => {})
}
function handleOpenAdbClick() {
  oneClickDebug()
}
function handleCloseAdbClick() {
  ElMessage.info('正在关闭 ADB，请稍候...')
  axios.post('/api/adb/close')
    .then(res => {
      const ok = res?.data?.code === 0 || res?.data?.success
      setTimeout(() => {
        if (ok) {
          ElMessage.success('ADB 调试已关闭')
          refreshAdbStatus()
        } else {
          ElMessage.error('ADB 关闭失败：' + (res?.data?.msg || '未知错误'))
        }
      }, 1200)
    })
    .catch(err => ElMessage.error('请求失败：' + (err?.message || '未知错误')))
}

// 重启设备的 in-flight 标记，防止重复点击
const restarting = ref(false)

async function handleRestartClick() {
  try {
    await ElMessageBox.confirm(
      '重启设备将断开所有网络连接(包括 SSH、网页、5G/4G/WiFi)，通常需要 60~120 秒恢复。\n\n' +
      '执行前请确认：\n' +
      '• 没有正在下载/上传的重要任务\n' +
      '• 已经保存好当前网页端的操作\n' +
      '• 当前能接受短暂离线\n\n' +
      '确认要重启设备吗？',
      '重启设备',
      {
        confirmButtonText: '确认重启',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--warning',
      },
    )
  } catch {
    return // 用户取消
  }
  restarting.value = true
  ElMessage.warning('正在重启设备，请耐心等待 60~120 秒...')
  axios.post('/api/sys/restart')
    .then(() => {
      // 重启后会断网，不再刷新状态; 60s 后再放开按钮
      setTimeout(() => { restarting.value = false }, 90000)
    })
    .catch(err => {
      restarting.value = false
      ElMessage.error('重启请求失败：' + (err?.message || '未知错误'))
    })
}

async function handleOpenFirewallClick() {
  try {
    await ElMessageBox.confirm(
      '【开启防火墙的好处】\n' +
      '• 阻止外部网络对内网设备的主动探测与攻击\n' +
      '• 拦截来自互联网的未授权访问请求\n' +
      '• 防止恶意流量进入局域网\n\n' +
      '【开启防火墙的弊端】\n' +
      '• 部分应用/游戏的 P2P 连接可能受限\n' +
      '• 自建服务（远程桌面/网站/NAS）需手动配置端口映射\n' +
      '• 部分内网穿透工具可能失效\n\n' +
      '确认要开启防火墙吗？',
      '开启防火墙',
      {
        confirmButtonText: '确认开启',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
  } catch {
    return
  }
  ElMessage.info('正在开启防火墙，请稍候...')
  axios.post('/api/sys/firewall/set', { enable: true })
    .then(res => {
      const ok = res?.data?.code === 0
      setTimeout(() => {
        if (ok) {
          ElMessage.success('防火墙已开启')
          refreshFirewallStatus()
        } else {
          ElMessage.error('防火墙开启失败：' + (res?.data?.msg || '未知错误'))
        }
      }, 1200)
    })
    .catch(err => ElMessage.error('请求失败：' + (err?.message || '未知错误')))
}

async function handleCloseFirewallClick() {
  try {
    await ElMessageBox.confirm(
      '【关闭防火墙的好处】\n' +
      '• 内网设备之间完全互通，文件共享/投屏更顺畅\n' +
      '• 部分 P2P 下载/联机游戏的连接质量更好\n' +
      '• 自建服务、内网穿透工具更易连通\n\n' +
      '【关闭防火墙的弊端】\n' +
      '• 局域网设备直接暴露在广域网风险下\n' +
      '• 外部扫描/爆破/漏洞利用的风险上升\n' +
      '• 一旦设备被入侵，无任何拦截\n\n' +
      '确认要关闭防火墙吗？',
      '关闭防火墙',
      {
        confirmButtonText: '确认关闭',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    )
  } catch {
    return
  }
  ElMessage.info('正在关闭防火墙，请稍候...')
  axios.post('/api/sys/firewall/set', { enable: false })
    .then(res => {
      const ok = res?.data?.code === 0
      setTimeout(() => {
        if (ok) {
          ElMessage.success('防火墙已关闭')
          refreshFirewallStatus()
        } else {
          ElMessage.error('防火墙关闭失败：' + (res?.data?.msg || '未知错误'))
        }
      }, 1200)
    })
    .catch(err => ElMessage.error('请求失败：' + (err?.message || '未知错误')))
}

function oneClickDebug() {
  ElMessage.info('正在执行操作，请稍候...')

  axios.post('/api/openadb')
    .then(res => {
      // 可以根据后端返回判断是否成功
      const success = res?.data?.code === 0 || res?.data?.success
      setTimeout(() => {
        if (success) {
          ElMessage.success('ADB 调试模式已开启')
          refreshAdbStatus()
        } else {
          ElMessage.error('ADB 开启失败，请重试')
        }
      }, 1500)
    })
    .catch(err => {
      ElMessage.error('请求失败：' + (err?.message || '未知错误'))
    })
}

// function oneClickDebugClose() {
//   ElMessage.info('正在执行操作，请稍候...')
//   callUbusBatch([closeAdbRequest])
//       .then((map) => {
//         // 一秒后执行刷新，确保后端状态更新
//         setTimeout(() => {
//             ElMessage.success('已关闭ADB调试模式')
//         }, 1500);
//       })
//       .catch((err) => {
//         ElMessage.error('请求失败：' + (err?.message || '未知错误'))
//       })
// }

// 短信转发
function smsForwardHandler() {
  ElNotification({
    title: '功能未实现',
    message: '短信转发功能尚未实现，敬请期待！',
    type: 'warning',
    duration: 5000,
  });
}

function netAmbrGetHandler() {
  axios.post('/api/net/ambr/get', { })
    .then((res) => {
      if (res.data.code !== 0) return;
      const data = res.data.data;
      netAmbr.value = {
        ...netAmbr.value,
        ...data,
        dl: { ...netAmbr.value.dl, ...data.dl },
        ul: { ...netAmbr.value.ul, ...data.ul },
      };
    })
}
function psmGetHandler() {
  // G5 Pro 无线接口为 MediaTek mt7990: ra0(2.4G) / rai0(5G) / rax0(6G)
  axios.post('/api/wifi/psm/get', { ifaces: ['ra0', 'rai0', 'rax0'], })
    .then((res) => {
      if (res.data.code !== 0) return;
      const data = res.data.data;
      wifiInfo.value.ra0 = data.ra0_psm;
      wifiInfo.value.rai0 = data.rai0_psm;
      wifiInfo.value.rax0 = data.rax0_psm;

      // 2.4G: ra0
      wifiInfo.value.wifiStatus24 = data.ra0_status === 'up';
      // 5G: rai0
      wifiInfo.value.wifiStatus5 = data.rai0_status === 'up';

      // psm
      wifiInfo.value.highPerformance = data.rai0_psm === 'off';
    })
}
function psmSetHandler(val:boolean){
  axios.post('/api/wifi/psm/set', {
    ifaces: ['ra0', 'rai0', 'rax0'],
    mode: val ? 'off' : 'on',
  }).then((res) => {
    psmGetHandler()
    ElMessage.success('WiFi已切换为:' + (val ? '高性能模式(据说会降低WiFi延迟)' : '省电模式'));
  });
}
function wifiStateSetHandler(iface:string, val:boolean){
  axios.post('/api/wifi/state/set', {
    ifaces: [iface],
    up: val,
  }).then((res) => {
    psmGetHandler()
    ElMessage.success((iface == 'wlan0' ? '2.4G' : ((iface == 'wlan2' ? '5G' : '其他'))) + '-WiFi已' + (val ? '开启' : '关闭'));
  });
}
function clampPercent(value: number) {
  if (Number.isNaN(value)) return 0;
  return Math.max(0, Math.min(100, value));
}

const totalCpuLoad = computed(() => {
  const idle = Number(deviceInfo.value.cpuinfo?.[0]?.idle ?? 100);
  return clampPercent(100 - idle);
});

const cpuCoreLoads = computed(() => {
  return [1, 2, 3, 4].map((index) => {
    const idle = Number(deviceInfo.value.cpuinfo?.[index]?.idle ?? 100);
    return {
      name: `核心${index}`,
      value: clampPercent(100 - idle),
    };
  });
});

const cpuPieStyle = computed(() => {
  const load = totalCpuLoad.value;
  const h = getComputedStyle(document.documentElement).getPropertyValue('--theme-primary-h').trim() || '201';
  const s = getComputedStyle(document.documentElement).getPropertyValue('--theme-primary-s').trim() || '100%';
  const l = getComputedStyle(document.documentElement).getPropertyValue('--theme-primary-l').trim() || '46%';
  return {
    '--pie-percent': load + '%',
    '--pie-fill': `hsl(${h} ${s} ${l} / 0.7)`,
    '--pie-track': `hsla(${h} ${s} ${l} / 0.13)`,
  };
});

function getTempPercent(temp: unknown): number {
  const n = Number(temp);
  if (Number.isNaN(n)) return 0;

  // 假设 0~100℃ 映射成 0~100%
  return Math.max(0, Math.min(100, n));
}

function getTempClass(temp: unknown): string {
  const n = Number(temp);
  if (Number.isNaN(n)) return 'normal';

  if (n >= 70) return 'danger';
  if (n >= 55) return 'warning';
  return 'normal';
}

function getTempText(temp: unknown): string {
  const n = Number(temp);
  if (Number.isNaN(n)) return '-';

  if (n >= 70) return '过热';
  if (n >= 55) return '偏高';
  return '正常';
}

onMounted(() => {
  fetchAllData();
  fetchAllData2();
  if (autoRefresh.value) {
    startAutoRefresh();
  }
  // 获取WiFi状态
  psmGetHandler();
  // 获取签约速率
  netAmbrGetHandler();
  // ADB 状态检测 + 周期轮询(每 5 秒)
  refreshAdbStatus();
  adbTimer = window.setInterval(refreshAdbStatus, 5000);
  // 防火墙状态检测 + 周期轮询(每 5 秒)
  refreshFirewallStatus();
  firewallTimer = window.setInterval(refreshFirewallStatus, 5000);
  // 温度状态检测(每 10 秒)
  refreshThermal();
  thermalTimer = window.setInterval(refreshThermal, 10000);
});

onUnmounted(() => {
  stopAutoRefresh();
  if (adbTimer !== null) {
    clearInterval(adbTimer);
    adbTimer = null;
  }
  if (firewallTimer !== null) {
    clearInterval(firewallTimer);
    firewallTimer = null;
  }
  if (thermalTimer !== null) {
    clearInterval(thermalTimer);
    thermalTimer = null;
  }
});
</script>

<style scoped>
/* 基础样式 */
.page {
  position: relative;
  color: var(--theme-text-color, white);
  min-height: 100vh;
  background:
    var(--theme-bg-image),
    linear-gradient(135deg,
      hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.85), 22%) 0%,
      hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.7), 33%) 50%,
      hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.6), 45%) 100%);
  background-size: cover, cover;
  background-repeat: no-repeat, no-repeat;
  background-position: center, center;
  background-attachment: fixed;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "HarmonyOS Sans SC", "Microsoft YaHei", "Noto Sans SC", "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

.page-overlay {
  position: absolute;
  inset: 0;
  background: var(--theme-overlay-color);
  pointer-events: none;
  z-index: 0;
  transition: background-color 0.3s ease;
}

/* 页面头部 */
.page-header {
  display: flex;
  flex-wrap: wrap;
  row-gap: 10px;
  justify-content: space-between;
  align-items: center;
  backdrop-filter: blur(var(--theme-blur-rate, 20px));
  border-radius: 16px;
  padding: 12px 16px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.08);
}


.title-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.title {
  font-size: 28px;
  font-weight: 700;
  color: var(--theme-text-color, #ffffff);
  margin: 0;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.status-dot.online {
  background: #48bb78;
}

.status-dot.offline {
  background: #e53e3e;
}

.status-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.8));
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* 控制区域 */
.controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 16px;
}

.ctrl-group {
  display: flex;
  flex-wrap: wrap;       /* 窄屏允许按钮换行, 避免溢出 */
  align-items: center;
  gap: 8px 10px;         /* row-gap col-gap */
  row-gap: 8px;
  white-space: nowrap;
}

.auto-refresh-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.9));
}

.checkbox-label input[type='checkbox'] {
  display: none;
}

.checkmark {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.6);
  border-radius: 4px;
  position: relative;
  transition: all 0.2s ease;
}

.checkbox-label input[type='checkbox']:checked + .checkmark {
  background: #4299e1;
  border-color: #4299e1;
}

.checkbox-label input[type='checkbox']:checked + .checkmark::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 12px;
  font-weight: bold;
}

.auto-refresh-controls select {
  padding: 4px 6px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.15);
  backdrop-filter: blur(var(--theme-blur-rate, 10px));
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.auto-refresh-controls select:focus {
  outline: none;
  border-color: var(--theme-btn-active-color, #4299e1);
}

.auto-refresh-controls select:disabled {
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.05);
  color: rgba(255, 255, 255, 0.4);
  cursor: not-allowed;
}

/* 按钮样式 */
.btn {
  padding: 5px 10px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--theme-btn-color, #409eff5d);
  color: white;
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.1);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(var(--theme-hover-y, -2px));
  box-shadow: 0 6px 16px rgba(66, 153, 225, 0.4);
  background: var(--theme-btn-active-color, #409eff);
}

.btn-danger {
  background: linear-gradient(135deg, #e53e3e, #c53030);
  color: white;
  box-shadow: 0 4px 12px rgba(229, 62, 62, 0.3);
}

.btn-danger:hover:not(:disabled) {
  transform: translateY(var(--theme-hover-y, -2px));
  box-shadow: 0 6px 16px rgba(229, 62, 62, 0.4);
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid transparent;
  border-top: 2px solid currentColor;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 状态页面 */
.loading,
.error,
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.15);
  backdrop-filter: blur(var(--theme-blur-rate, 20px));
  border-radius: 16px;
  padding: 40px;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.loading-spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e2e8f0;
  border-top: 4px solid #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

.error-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.loading p,
.error h3,
.error p,
.empty h3,
.empty p {
  color: rgba(255, 255, 255, 0.9);
  margin: 0;
}

.error h3,
.empty h3 {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 8px;
}

/* 内容区域 - 等宽网格 */
.content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  justify-content: center;
  gap: 24px;
  align-items: stretch; /* 卡片等高 */
}

/* 卡片样式 */
.card {
  backdrop-filter: blur(var(--theme-blur-rate, 20px));
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  display: flex;
  flex-direction: column;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.06);
}

.card:hover {
  transform: translateY(var(--theme-hover-y, -4px));
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
}

.hd {
  display: flex;
  gap: 10px;
}
.hd img {
  width: 24px;
}

.card-header {
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.1);
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  font-size: 17px;
  font-weight: 700;
  color: var(--theme-text-color, #ffffff);
  margin: 0;
  letter-spacing: 0.2px;
}

.card-tags {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}

.card-content {
  padding: 16px;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
}

/* 设备信息卡片 */
.device-info-card {
  grid-column: 1 / -1; /* 占据整行 */
}

/* 操作卡片 */
.device-actions-card {
  grid-column: 1 / -1; /* 占据整行 */
}

.device-stats {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 24px;
  align-items: stretch;
}

.device-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.health-card {
  position: relative;          /* 相对定位，避免被父元素限制 */
  z-index: 1;                  /* 确保浮动在上层显示 */
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  flex: 1 1 auto;              /* 设备信息内 CPU/温度/内存 三个子卡等高 */
  min-height: 0;
}

.health-card:hover {
  transform: translateY(var(--theme-hover-y, -3px));
  box-shadow: 0 6px 12px rgba(0,0,0,0.15);
}

.device-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.6px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.62));
  text-transform: uppercase;
  margin-bottom: 10px;
}

.device-values {
  display: flex;
  gap: 16px;
}

.load-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.load-label {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.4px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.6));
  text-transform: uppercase;
}

.load-value {
  font-size: 15px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
}

.memory-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.memory-bar {
  position: relative;
  height: 20px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.1);
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.memory-fill {
  height: 100%;
  background: linear-gradient(
    90deg,
    #62718abb 0%,
    #68d391bb 70%,
    #63b3edbb 100%
  );
  border-radius: 10px;
  transition: width 0.3s ease;
}

.memory-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 11px;
  font-weight: 600;
  color: #ffffff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.memory-details {
  display: flex;
  align-items: baseline;
  gap: 4px;
  font-size: 15px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.88));
}

.memory-used {
  font-weight: 600;
}

.memory-separator {
  color: var(--theme-text-color, rgba(255, 255, 255, 0.5));
}

.memory-total {
  color: var(--theme-text-color, rgba(255, 255, 255, 0.7));
}

.temp-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.temp-value {
  font-size: 16px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
}

.temp-bar {
  height: 16px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.1);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.temp-fill {
  height: 100%;
  background: linear-gradient(
    90deg,
    #62718abb 0%,
    #68d391bb 70%,
    #63b3edbb 100%
  );
  border-radius: 8px;
  transition: width 0.3s ease;
}

.uptime-value {
  font-size: 20px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
}

/* 信息网格：默认双列，14px/11px 三档层级紧凑布局 */
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 18px;
  row-gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 0;
}

/* 网络信息 / 频段 / 设备标识 三组小标题——横跨两列，置于分组顶部 */
.info-section-title {
  grid-column: 1 / -1;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.8px;
  text-transform: uppercase;
  color: var(--theme-text-color, rgba(140, 210, 255, 0.78));
  margin: 14px 0 0 0;
  padding: 0 0 4px 0;
  border-bottom: 1px dashed rgba(140, 210, 255, 0.18);
}
.info-section-title:first-child {
  margin-top: 0;
}

.info-item .label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.6px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.62));
  text-transform: uppercase;
  line-height: 1.2;
}

.info-item .value {
  font-size: 14px;
  font-weight: 600;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.96));
  line-height: 1.4;
  word-break: break-all;
  overflow-wrap: anywhere;
  white-space: normal;
}

/* 标签样式 */
.tag {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tag.success {
  background: rgba(54, 237, 69, 0.2);
  color: #75f655;
  border: 1px solid rgba(54, 237, 69, 0.3);
}

.tag.warning {
  background: rgba(237, 137, 54, 0.2);
  color: #f6ad55;
  border: 1px solid rgba(237, 137, 54, 0.3);
}

.tag.danger {
  background: rgba(229, 62, 62, 0.2);
  color: #fc8181;
  border: 1px solid rgba(229, 62, 62, 0.3);
}

.tag.excellent {
  background: rgba(72, 187, 120, 0.22);
  color: #7ee787;
  border: 1px solid rgba(72, 187, 120, 0.35);
}

.tag.good {
  background: rgba(56, 189, 248, 0.18);
  color: #7dd3fc;
  border: 1px solid rgba(56, 189, 248, 0.34);
}

.tag.fair {
  background: rgba(237, 137, 54, 0.2);
  color: #f6ad55;
  border: 1px solid rgba(237, 137, 54, 0.3);
}

.tag.poor {
  background: rgba(229, 62, 62, 0.2);
  color: #fc8181;
  border: 1px solid rgba(229, 62, 62, 0.3);
}

.tag.unknown {
  background: rgba(148, 163, 184, 0.16);
  color: #cbd5e1;
  border: 1px solid rgba(148, 163, 184, 0.26);
}

/* 信号进度条 */
.signal-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 16px;
}

.signal-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.signal-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.signal-label-help {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
}

.signal-help-trigger {
  width: 18px;
  height: 18px;
  padding: 0;
  border: 1px solid rgba(125, 211, 252, 0.45);
  border-radius: 50%;
  background: rgba(56, 189, 248, 0.12);
  color: #7dd3fc;
  cursor: pointer;
  font-size: 13px;
  font-weight: 800;
  line-height: 18px;
  text-align: center;
  transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}

.signal-help-trigger:hover,
.signal-help-trigger[aria-expanded='true'] {
  background: rgba(56, 189, 248, 0.24);
  border-color: rgba(125, 211, 252, 0.7);
  color: #dff6ff;
}

.signal-help-panel {
  padding: 10px 12px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 8px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.18);
  color: rgba(255, 255, 255, 0.82);
  font-size: 12px;
  line-height: 1.55;
}

.signal-help-title {
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
}

.signal-help-desc {
  margin-top: 4px;
}

.signal-help-ranges {
  display: grid;
  gap: 4px;
  margin-top: 8px;
}

.signal-help-ranges div {
  display: flex;
  align-items: center;
  gap: 6px;
}

.signal-help-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex: 0 0 auto;
}

.signal-help-dot.excellent {
  background: #68d391;
}

.signal-help-dot.good {
  background: #38bdf8;
}

.signal-help-dot.fair {
  background: #f6ad55;
}

.signal-help-dot.poor {
  background: #fc8181;
}

.signal-status {
  flex: 0 0 auto;
  min-width: 38px;
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.35;
  text-align: center;
  white-space: nowrap;
}

.signal-status.excellent {
  background: rgba(72, 187, 120, 0.18);
  color: #7ee787;
}

.signal-status.good {
  background: rgba(56, 189, 248, 0.16);
  color: #7dd3fc;
}

.signal-status.fair {
  background: rgba(237, 137, 54, 0.18);
  color: #f6ad55;
}

.signal-status.poor {
  background: rgba(229, 62, 62, 0.18);
  color: #fc8181;
}

.signal-status.unknown {
  background: rgba(148, 163, 184, 0.14);
  color: #cbd5e1;
}

.progress-bar {
  position: relative;
  height: 24px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.1);
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #64748bbb 0%, #38bdf8bb 100%);
  border-radius: 12px;
  transition: width 0.3s ease;
  position: relative;
}

.progress-fill.excellent {
  background: linear-gradient(90deg, #2f855abb 0%, #68d391dd 100%);
}

.progress-fill.good {
  background: linear-gradient(90deg, #0f766ebb 0%, #38bdf8dd 100%);
}

.progress-fill.fair {
  background: linear-gradient(90deg, #8a5a1fbb 0%, #f6ad55dd 100%);
}

.progress-fill.poor {
  background: linear-gradient(90deg, #7f1d1dbb 0%, #fc8181dd 100%);
}

.progress-fill.unknown {
  background: linear-gradient(90deg, #475569bb 0%, #94a3b8bb 100%);
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: 600;
  color: var(--theme-text-color, #ffffff);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  z-index: 1;
}

/* 顶部网络信息卡片中的信号条图标 */
.signal-strength-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.signal-strength-num {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", Menlo, Consolas, monospace;
  font-variant-numeric: tabular-nums;
  font-size: 13px;
  font-weight: 700;
}
.signal-bars {
  display: inline-flex;
  align-items: flex-end;
  gap: 3px;
}

.bar {
  width: 5px;
  height: 6px;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 2px;
  transition: height 0.2s ease, background 0.2s ease;
}

.full-bar {
  height: 18px;
}

.bar:nth-child(1) {
  height: 6px;
}
.bar:nth-child(2) {
  height: 9px;
}
.bar:nth-child(3) {
  height: 12px;
}
.bar:nth-child(4) {
  height: 15px;
}
.bar:nth-child(5) {
  height: 18px;
}

.bar.active:nth-child(1) {
  background: #68d391;
}
.bar.active:nth-child(2) {
  background: #68d391;
}
.bar.active:nth-child(3) {
  background: #68d391;
}
.bar.active:nth-child(4) {
  background: #68d391;
}
.bar.active:nth-child(5) {
  background: #68d391;
}

.battery {
  position: relative;
  width: 30px;
  height: 14px;
  border: 2px solid #ffffffc4;
  border-radius: 3px;
  box-sizing: border-box;
  display: inline-block;
}

.battery-head {
  position: absolute;
  right: -4px;
  top: 1px;
  width: 3px;
  height: 8px;
  background: #ffffffc4;
  border-radius: 0 2px 2px 0;
}

.battery-level {
  height: 100%;
  background: #40d67a; /* 默认绿色 */
  transition: width 0.3s ease, background 0.3s ease;
  border-radius: 1px;
}

.battery-charging {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -55%) scale(1.8) rotate(20deg);
  font-size: 10px;
  color: #fff200;
  display: none; /* 默认不显示 */
}

/* 电量低时变红 */
.battery.low .battery-level {
  background: #f56565;
}

/* 充电状态显示闪电 */
.battery.charging .battery-charging {
  display: block;
}

/* 接口区域 */
.interface-section {
  margin-bottom: 16px;
}

.interface-section:last-child {
  margin-bottom: 0;
}

.interface-section h4 {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.8px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.75));
  text-transform: uppercase;
  margin: 0 0 10px 0;
  padding-bottom: 6px;
  border-bottom: 1px solid rgba(calc(var(--theme-text-gray, 255) * 1), calc(var(--theme-text-gray, 255) * 1), calc(var(--theme-text-gray, 255) * 1), 0.18);
}

/* 网络接口状态网格布局 */
.interface-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}
.interface-grid > .interface-section {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
}
.interface-grid > .interface-section > .info-grid-compact {
  flex: 1 1 auto;
  align-content: start;
}

.info-grid-compact {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  column-gap: 16px;
  row-gap: 10px;
}

.info-grid-compact .info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-grid-compact .info-item .label {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.55));
  text-transform: uppercase;
  line-height: 1.2;
}

.info-grid-compact .info-item .value {
  font-size: 13px;
  font-weight: 600;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
  line-height: 1.4;
  word-break: break-all;
}

/* 流量统计 */
.traffic-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
}

.traffic-item {
  text-align: center;
  padding: 16px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.1);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: center;
  gap: 6px;
  min-height: 78px; /* 与单行大字号速度 tile 视觉等高 */
}

.traffic-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.62));
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.traffic-value {
  font-size: 17px;
  font-weight: 700;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
  font-variant-numeric: tabular-nums;
  line-height: 1.3;
}

.traffic-value.upload {
  color: #fc8181;
}

.traffic-value.download {
  color: #68d391;
}

/* 用量（上传/下载）双行紧凑：与单行速度卡视觉等高
   替代原来两行高度撑高造成 8 tile 高度不齐的问题 */
.traffic-value-stack {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.25;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
  font-variant-numeric: tabular-nums;
}
.traffic-stack-row {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  white-space: nowrap;
}
.traffic-stack-k {
  font-size: 10px;
  font-weight: 600;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.55));
  letter-spacing: 0.3px;
  text-transform: uppercase;
}
.traffic-stack-v {
  font-size: 15px;
  font-weight: 800;
  color: var(--theme-text-color, #ffffff);
  font-variant-numeric: tabular-nums;
}
.traffic-value.upload .traffic-stack-v { color: #fc8181; }
.traffic-value.download .traffic-stack-v { color: #68d391; }

/* 响应式设计 */
@media (max-width: 768px) {
  .page {
    padding: 12px;
  }

  .page-header {
    flex-direction: column;
    gap: 16px;
    padding: 16px;
  }

  .title {
    font-size: 24px;
  }

  .controls {
    flex-direction: column;
    gap: 12px;
    width: 100%;
  }

  .auto-refresh-controls {
    justify-content: center;
  }

  .content {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .interface-grid {
    grid-template-columns: 1fr;
  }

  .info-grid-compact {
    grid-template-columns: 1fr;
  }

  .traffic-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  /* 手机端 info-item 水平排列：标签左、值右，避免值被压成竖排字符 */
  .info-item {
    flex-direction: row;
    justify-content: space-between;
    align-items: baseline;
    gap: 8px;
    padding: 0;
  }
  .info-item .label {
    font-size: 10px;
    flex-shrink: 0;
  }
  .info-item .value {
    font-size: 12px;
    line-height: 1.4;
    text-align: right;
    word-break: normal;
    overflow-wrap: normal;
  }
  .card-content {
    padding: 12px;
  }
}

@media (max-width: 1100px) {
  .page {
    padding: 8px;
  }

  .card-content {
    padding: 16px;
  }

  .device-stats {
    grid-template-columns: 1fr;
  }

  .traffic-item {
    padding: 12px;
  }

  .traffic-value {
    font-size: 16px;
  }
}

.mytable{caption-side: bottom;}
.mytable{border-collapse: collapse;}
.mytable,tr,td{border: 1px solid rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.35);font-size: 14px;text-align: center;color: var(--theme-text-color, rgba(255,255,255,0.85));}
.dbmstyle{color: rgb(104, 211, 145)}

.hp {
  color: #d97706;
  font-weight: 600;
}
.psm {
  color: #059669;
  font-weight: 600;
}

.btn-disabled {
  background: #9ca3af;
  border-color: #9ca3af;
  color: #fff;
  cursor: not-allowed;
  opacity: 0.75;
}

/* 表格手机端横向滚动 */
.table-wrapper {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.table-wrapper .mytable {
  min-width: 720px;
  width: 100%;
}

.child {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1600px;
  margin: 0 auto;
  padding: 16px;
  box-sizing: border-box;
}

.top-cards {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  width: 100%;
  align-items: stretch; /* NR 5G 信号 与 5G 载波信息 两卡等高 */
}

.top-cards .card {
  width: 100%;
  min-width: 0;
}

@media (max-width: 900px) {
  .top-cards {
    grid-template-columns: 1fr;
  }
}

/* 网络信息 / 流量统计 / 接口状态：三卡片等宽网格（整体大小一致） */
.bottom-cards {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
  width: 100%;
  align-items: stretch; /* 等宽 + 等高：三卡片视觉上同一大小 */
}

/* 手机 / iPad（含横屏）：网络信息 / 流量统计 / 接口状态 三卡竖排单列，不并排 */
@media (max-width: 1200px) {
  .bottom-cards {
    grid-template-columns: 1fr;
  }
}

.bottom-cards > .card {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.bottom-cards > .card > .card-content {
  flex: 1 1 auto;
}

.cpu-health-card {
  min-width: 0;
}

.cpu-health-layout {
  display: grid;
  grid-template-columns: 150px 1fr;
  align-items: center;
}

.cpu-pie-box {
  display: flex;
  align-items: center;
  justify-content: center;
}

.cpu-pie {
  position: relative;
  width: 115px;
  height: 115px;
  border-radius: 50%;
  background: conic-gradient(
    var(--pie-fill) 0 var(--pie-percent),
    var(--pie-track) var(--pie-percent) 100%
  );
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.08),
    0 16px 34px rgba(0, 0, 0, 0.26);
}

.cpu-pie-inner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 83px;
  height: 83px;
  border-radius: 50%;
  background: hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.8), 28%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.cpu-pie-value {
  font-size: 30px;
  font-weight: 900;
  line-height: 1;
  color: #fff;
}

.cpu-pie-text {
  margin-top: 8px;
  font-size: 13px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.62));
}

.cpu-core-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.cpu-core-card {
  min-width: 0;
  padding: 10px 12px;
  padding-bottom: 16px;
  border-radius: 14px;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.045);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.cpu-core-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}

.cpu-core-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.62));
}

.cpu-core-value {
  font-size: 16px;
  font-weight: 900;
  color: var(--theme-text-color, #fff);
}

.cpu-core-bar {
  height: 8px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.12);
}

.cpu-core-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg,
    hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.8), 55%),
    hsl(calc(var(--theme-primary-h, 201) * 1 + 120), calc(var(--theme-primary-s, 100%) * 0.8), 55%));
  transition: width 0.25s ease;
}

.health-card {
  min-width: 0;
  padding: 18px;
  border-radius: 18px;
  background:
    radial-gradient(circle at top left,
      rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.12), transparent 38%),
    rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.045);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.health-title {
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.68));
  margin-bottom: 16px;
}

.temp-gauges {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  align-items: stretch;
}

.temp-gauge {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  min-width: 0;
  width: 100%;
}

.temp-ring {
  --ring-color: hsl(calc(var(--theme-primary-h, 201) * 1 + 120), calc(var(--theme-primary-s, 100%) * 0.8), 55%);

  position: relative;
  width: 104px;
  height: 104px;
  aspect-ratio: 1 / 1;
  flex: 0 0 auto;
  border-radius: 50%;
  background:
    conic-gradient(
      var(--ring-color) 0 var(--percent),
      rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.13) var(--percent) 100%
    );
  transition: background 0.25s ease;
}

.temp-ring.normal {
  --ring-color: hsl(calc(var(--theme-primary-h, 201) * 1 + 120), calc(var(--theme-primary-s, 100%) * 0.8), 55%);
}

.temp-ring.warning {
  --ring-color: #ffd166;
}

.temp-ring.danger {
  --ring-color: #ff6b6b;
}

.temp-ring-inner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.8), 28%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* 暗色模式支持 */
@media (prefers-color-scheme: dark) {
  .page {
    color: var(--theme-text-color, white);
    background:
      var(--theme-bg-image),
      linear-gradient(135deg,
        hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.8), 8%) 0%,
        hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.6), 16%) 50%,
        hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.5), 24%) 100%);
    background-size: cover, cover;
    background-repeat: no-repeat, no-repeat;
    background-position: center, center;
    background-attachment: fixed;
  }

  .page-header,
  .card,
  .loading,
  .error,
  .empty {
    background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.06);
    border-color: rgba(255, 255, 255, 0.1);
  }

  .title,
  .card-header h3,
  .info-item .value,
  .traffic-value {
    color: var(--theme-text-color, #f1f5f9);
  }

  .status-text,
  .info-item .label,
  .traffic-label {
    color: var(--theme-text-color, #cbd5e1);
  }

  .card-header {
    background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.08);
  }

  .interface-section h4 {
    color: #f1f5f9;
    border-bottom-color: rgba(255, 255, 255, 0.1);
  }
}


.temp-ring-inner strong {
  font-size: 22px;
  line-height: 1;
  color: var(--theme-text-color, #fff);
}

.temp-ring-inner span {
  margin-top: 6px;
  font-size: 12px;
  color: var(--theme-text-color, rgba(255,255,255,0.6));
}

.temp-state {
  padding: 4px 14px;
  min-width: 56px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
  text-align: center;
  background: rgba(255,255,255,0.08);
}

.temp-state.normal {
  color: #63e6be;
  background: rgba(99, 230, 190, 0.14);
}

.temp-state.warning {
  color: #ffd166;
  background: rgba(255, 209, 102, 0.14);
}

.temp-state.danger {
  color: #ff6b6b;
  background: rgba(255, 107, 107, 0.14);
}

.memory-health-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.memory-big {
  font-size: 46px;
  font-weight: 900;
  line-height: 1;
  color: var(--theme-text-color, #ffffff);
}

.memory-detail {
  margin-top: 8px;
  font-size: 18px;
  font-weight: 800;
  color: var(--theme-text-color, rgba(255,255,255,0.9));
}

.memory-detail span {
  font-size: 14px;
  color: var(--theme-text-color, rgba(255,255,255,0.55));
}

.memory-stack {
  margin-top: 18px;
  height: 16px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(calc(var(--theme-primary-r, 0) * 1), calc(var(--theme-primary-g, 97) * 1), calc(var(--theme-primary-b, 255) * 1), 0.12);
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.05);
}

.memory-stack-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg,
    hsl(calc(var(--theme-primary-h, 201) * 1), calc(var(--theme-primary-s, 100%) * 0.8), 55%),
    hsl(calc(var(--theme-primary-h, 201) * 1 + 120), calc(var(--theme-primary-s, 100%) * 0.8), 55%));
}

.memory-caption {
  margin-top: 10px;
  font-size: 12px;
  color: var(--theme-text-color, rgba(255,255,255,0.48));
}

@media (max-width: 100px) {
  .cpu-health-layout {
    grid-template-columns: 1fr;
  }

  .cpu-core-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

/* ===== 科技感增强（玻璃拟态 + 霓虹） ===== */
.page::after {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background-image:
    linear-gradient(rgba(120, 200, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(120, 200, 255, 0.05) 1px, transparent 1px);
  background-size: 32px 32px;
  -webkit-mask-image: radial-gradient(circle at 50% 25%, black, transparent 82%);
  mask-image: radial-gradient(circle at 50% 25%, black, transparent 82%);
}

.card {
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.10), rgba(255, 255, 255, 0.04));
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(140, 210, 255, 0.22);
  border-radius: 18px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.08);
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
}
.card:hover {
  transform: translateY(-2px);
  border-color: rgba(140, 210, 255, 0.45);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45), 0 0 24px rgba(80, 190, 255, 0.25), inset 0 1px 0 rgba(255, 255, 255, 0.10);
}
.card-header {
  position: relative;
}
.card-header .hd,
.card-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: #8be9ff;
  text-shadow: 0 0 16px rgba(90, 200, 255, 0.45);
}
.card-header::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: linear-gradient(90deg, rgba(120, 200, 255, 0.6), transparent);
}

/* 卡片头图标：统一霓虹线性图标，深色卡片上始终可见 */
.hd-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  flex: 0 0 auto;
  color: #8be9ff;
  background: linear-gradient(135deg, rgba(90, 200, 255, 0.18), rgba(90, 200, 255, 0.06));
  border: 1px solid rgba(120, 210, 255, 0.4);
  border-radius: 9px;
  box-shadow: inset 0 0 8px rgba(90, 200, 255, 0.18), 0 0 8px rgba(90, 200, 255, 0.22);
  filter: drop-shadow(0 0 4px rgba(90, 200, 255, 0.4));
}
.hd-icon svg {
  width: 19px;
  height: 19px;
  display: block;
}

.btn {
  border-radius: 10px;
  border: 1px solid rgba(140, 210, 255, 0.25);
  background: rgba(255, 255, 255, 0.06);
  color: var(--theme-text-color, #eaf6ff);
  padding: 6px 14px;
  font-weight: 600;
  letter-spacing: 0.3px;
  transition: all 0.2s ease;
  cursor: pointer;
}
.btn:hover:not(:disabled) {
  border-color: rgba(140, 210, 255, 0.6);
  box-shadow: 0 0 14px rgba(90, 200, 255, 0.35);
  transform: translateY(-1px);
}
.btn-primary {
  background: linear-gradient(135deg, #2bd2ff, #2b7bff);
  border-color: transparent;
  color: #04121f;
  box-shadow: 0 0 16px rgba(43, 200, 255, 0.45);
}
.btn-primary:hover:not(:disabled) {
  box-shadow: 0 0 24px rgba(43, 200, 255, 0.7);
}
.btn-danger {
  background: linear-gradient(135deg, #ff5b6e, #ff2d55);
  border-color: transparent;
  color: #1a0307;
  box-shadow: 0 0 16px rgba(255, 70, 100, 0.45);
}
.btn-success {
  background: linear-gradient(135deg, #2bff9e, #18c97a);
  border-color: transparent;
  color: #042016;
  box-shadow: 0 0 16px rgba(40, 230, 150, 0.4);
}
.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
}
.btn-warning {
  background: linear-gradient(135deg, #ffb84d, #ff7e2d);
  border-color: transparent;
  color: #2a1500;
  box-shadow: 0 0 16px rgba(255, 160, 60, 0.45);
}
.btn-warning:hover:not(:disabled) {
  box-shadow: 0 0 24px rgba(255, 160, 60, 0.7);
}
.btn-disabled,
.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  box-shadow: none;
  filter: grayscale(0.4);
}

/* 系统动作按钮(重启/防火墙开关)：与 ADB 按钮共用紧凑样式 */
.sys-action-toggle {
  padding: 6px 10px;
  font-size: 13px;
  white-space: nowrap;
}

/* ADB 开关按钮：紧凑 + 允许单字符被压缩, 防止窄屏溢出 */
.adb-toggle {
  padding: 6px 10px;
  font-size: 13px;
  white-space: nowrap;
}

/* 顶部 6 按钮：默认桌面一排 6 个；窄屏 ≤768px 自动 2×3 */
.ctrl-row > button {
  flex: 0 0 calc((100% - 50px) / 6);
  min-width: 0;
  justify-content: center;
}

@media (max-width: 768px) {
  .ctrl-row > button {
    flex: 0 0 calc((100% - 20px) / 3);
  }
}

.traffic-value {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.5px;
  text-shadow: 0 0 12px rgba(90, 200, 255, 0.3);
}
.traffic-value.upload { color: #6be4ff; }
.traffic-value.download { color: #7dffb0; }

.signal-bars .bar.active {
  background: linear-gradient(180deg, #8be9ff, #2b7bff);
  box-shadow: 0 0 8px rgba(90, 200, 255, 0.6);
}

.controls {
  background: rgba(255, 255, 255, 0.04);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(140, 210, 255, 0.15);
  border-radius: 14px;
  padding: 10px 14px;
}
.net-tag {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  color: #8be9ff;
  border: 1px solid rgba(140, 210, 255, 0.3);
  background: rgba(90, 200, 255, 0.08);
}

/* AMBR 大卡：2x2 网格布局，下行/上行/QCI/SIM卡号 整齐对齐 */
.ambr-hero {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 18px;
  align-items: stretch; /* 4 tile 视觉等高，关键 */
}
.ambr-tile {
  flex: initial;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03));
  border: 1px solid rgba(140, 210, 255, 0.22);
  position: relative;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  min-width: 0;
  min-height: 76px; /* 4 tile 等高（与 QCI/Score 视觉对齐） */
}
.ambr-tile:hover {
  transform: translateY(-2px);
}
.ambr-tile.down {
  box-shadow: inset 0 0 24px rgba(45, 226, 160, 0.12), 0 0 18px rgba(45, 226, 160, 0.10);
  border-color: rgba(45, 226, 160, 0.35);
}
.ambr-tile.down:hover {
  box-shadow: inset 0 0 24px rgba(45, 226, 160, 0.18), 0 0 24px rgba(45, 226, 160, 0.22);
}
.ambr-tile.up {
  box-shadow: inset 0 0 24px rgba(90, 200, 255, 0.12), 0 0 18px rgba(90, 200, 255, 0.10);
  border-color: rgba(90, 200, 255, 0.35);
}
.ambr-tile.up:hover {
  box-shadow: inset 0 0 24px rgba(90, 200, 255, 0.18), 0 0 24px rgba(90, 200, 255, 0.22);
}
.ambr-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 12px;
  flex: 0 0 auto;
}
.ambr-tile.down .ambr-ico {
  color: #2bff9e;
  background: rgba(45, 226, 160, 0.14);
  filter: drop-shadow(0 0 8px rgba(45, 226, 160, 0.6));
}
.ambr-tile.up .ambr-ico {
  color: #5ac8ff;
  background: rgba(90, 200, 255, 0.14);
  filter: drop-shadow(0 0 8px rgba(90, 200, 255, 0.6));
}
.ambr-ico svg {
  width: 26px;
  height: 26px;
  display: block;
}
.ambr-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.ambr-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.6px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.65));
  text-transform: uppercase;
  line-height: 1.2;
}
.ambr-val {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  font-size: 22px;
  font-weight: 800;
  line-height: 1.15;
}
.ambr-tile.down .ambr-val {
  color: #6bffb0;
  text-shadow: 0 0 14px rgba(45, 226, 160, 0.45);
}
.ambr-tile.up .ambr-val {
  color: #8be9ff;
  text-shadow: 0 0 14px rgba(90, 200, 255, 0.45);
}
.ambr-unit {
  font-size: 12px;
  font-weight: 700;
  margin-left: 3px;
  opacity: 0.75;
}
.ambr-qci {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(140, 210, 255, 0.2);
  min-width: 0;
  min-height: 76px; /* 与 .ambr-tile 等高 */
  gap: 4px;
}
/* SIM 卡号 tile：与 .ambr-qci 视觉对齐，金色突出 */
.ambr-sim {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 16px;
  border-radius: 14px;
  background: rgba(255, 216, 107, 0.06);
  border: 1px solid rgba(255, 216, 107, 0.28);
  min-width: 0;
  gap: 4px;
  text-align: center;
  word-break: break-all;
  overflow-wrap: anywhere;
}
.ambr-qci-val {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-size: 24px;
  font-weight: 800;
  color: #fff;
  line-height: 1.15;
  text-shadow: 0 0 14px rgba(90, 200, 255, 0.4);
}
/* 信号强度 tile：与 .ambr-qci 视觉对齐，按 RSRP 强度切换 4 档主题色 */
.ambr-signal {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(140, 210, 255, 0.22);
  min-width: 0;
  min-height: 76px; /* 与 .ambr-tile 等高 */
  gap: 4px;
  text-align: center;
  transition: background 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
}
.ambr-signal.grade-excellent {
  background: rgba(45, 226, 160, 0.10);
  border-color: rgba(45, 226, 160, 0.42);
  box-shadow: inset 0 0 18px rgba(45, 226, 160, 0.10);
}
.ambr-signal.grade-good {
  background: rgba(90, 200, 255, 0.08);
  border-color: rgba(90, 200, 255, 0.40);
}
.ambr-signal.grade-fair {
  background: rgba(255, 200, 100, 0.08);
  border-color: rgba(255, 200, 100, 0.40);
}
.ambr-signal.grade-poor {
  background: rgba(255, 90, 90, 0.08);
  border-color: rgba(255, 90, 90, 0.42);
}
.ambr-signal-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.ambr-signal-bars {
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 18px;
}
.ambr-sig-bar {
  width: 4px;
  background: rgba(255, 255, 255, 0.14);
  border-radius: 1.5px;
  transition: background 0.3s ease, box-shadow 0.3s ease;
}
.ambr-sig-bar:nth-child(1) { height: 5px; }
.ambr-sig-bar:nth-child(2) { height: 9px; }
.ambr-sig-bar:nth-child(3) { height: 13px; }
.ambr-sig-bar:nth-child(4) { height: 16px; }
.ambr-sig-bar:nth-child(5) { height: 19px; }
.ambr-signal.grade-excellent .ambr-sig-bar.active {
  background: #2de2a0;
  box-shadow: 0 0 6px rgba(45, 226, 160, 0.7);
}
.ambr-signal.grade-good .ambr-sig-bar.active {
  background: #5ac8ff;
  box-shadow: 0 0 6px rgba(90, 200, 255, 0.7);
}
.ambr-signal.grade-fair .ambr-sig-bar.active {
  background: #ffc864;
  box-shadow: 0 0 6px rgba(255, 200, 100, 0.6);
}
.ambr-signal.grade-poor .ambr-sig-bar.active {
  background: #ff5a5a;
  box-shadow: 0 0 6px rgba(255, 90, 90, 0.6);
}
.ambr-signal-dbm {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  font-size: 18px;
  font-weight: 800;
  line-height: 1;
  color: #fff;
}
.ambr-signal.grade-excellent .ambr-signal-dbm { color: #6bffb0; text-shadow: 0 0 10px rgba(45, 226, 160, 0.45); }
.ambr-signal.grade-good .ambr-signal-dbm { color: #8be9ff; text-shadow: 0 0 10px rgba(90, 200, 255, 0.45); }
.ambr-signal.grade-fair .ambr-signal-dbm { color: #ffd86b; text-shadow: 0 0 10px rgba(255, 216, 107, 0.45); }
.ambr-signal.grade-poor .ambr-signal-dbm { color: #ff7a7a; text-shadow: 0 0 10px rgba(255, 90, 90, 0.45); }
/* SINR 副指标：等宽数字，标签更暗；评级 tile 用 RSSI 决定主题色，SINR 仅作信息补充 */
.ambr-signal-snr {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
  color: rgba(234, 246, 255, 0.82);
  letter-spacing: 0.3px;
  margin-top: 2px;
}
.ambr-signal-snr .snr-label {
  color: rgba(255, 255, 255, 0.5);
  margin-right: 5px;
  font-weight: 500;
  font-size: 11px;
  letter-spacing: 0.5px;
}
.ambr-signal-rating {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.6px;
  opacity: 0.85;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.72));
}
/* === 网络综合评分 tile === */
.ambr-score-value {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  font-size: 30px;
  font-weight: 900;
  line-height: 1.05;
  text-shadow: 0 0 12px rgba(90, 200, 255, 0.35);
}
.ambr-score-detail {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px 12px;
  width: 100%;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.3px;
  color: rgba(234, 246, 255, 0.75);
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  margin-top: 2px;
}
.ambr-score-detail span {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  line-height: 1.05;
  min-width: 0;
}
.ambr-score-detail em {
  font-style: normal;
  font-weight: 800;
  font-size: 12px;
}
.ambr-score-detail i {
  font-style: normal;
  font-weight: 500;
  font-size: 9.5px;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: 0.4px;
  margin-top: 1px;
}
.ambr-signal.grade-excellent .ambr-score-value { text-shadow: 0 0 12px rgba(45, 226, 160, 0.45); }
.ambr-signal.grade-good      .ambr-score-value { text-shadow: 0 0 12px rgba(90, 200, 255, 0.45); }
.ambr-signal.grade-fair      .ambr-score-value { text-shadow: 0 0 12px rgba(255, 200, 100, 0.45); }
.ambr-signal.grade-poor      .ambr-score-value { text-shadow: 0 0 12px rgba(255, 90, 90, 0.45); }
/* QCI 副标题：让 QCI tile 与评分 tile 信息密度对齐（避免「9」独占一格太空） */
.ambr-qci-sub {
  font-size: 10px;
  letter-spacing: 0.3px;
  color: rgba(234, 246, 255, 0.65);
  font-weight: 500;
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}
.ambr-iccid {
  flex: 0 1 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 18px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(140, 210, 255, 0.2);
  min-width: 150px;
  gap: 4px;
}
/* SIM 卡号数值（hero 第二行右侧，金色，适配长号码自动换行） */
.ambr-sim-val {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", "Roboto Mono", Menlo, Consolas, "Liberation Mono", monospace;
  font-variant-numeric: tabular-nums;
  font-size: 16px;
  font-weight: 800;
  color: #ffd86b;
  line-height: 1.25;
  text-shadow: 0 0 12px rgba(255, 216, 107, 0.45);
  letter-spacing: 0.2px;
  word-break: break-all;
  overflow-wrap: anywhere;
  text-align: center;
}
/* conntrack 连接状态子区块（位于流量统计卡片内） */
.conn-track-block {
  margin-top: 14px;
}
.ct-divider {
  border-top: 1px dashed rgba(255, 255, 255, 0.25);
  margin: 6px 0 10px;
}
.ct-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.ct-header .hd-icon {
  display: inline-flex;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.9));
}
.ct-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--theme-text-color, #ffffff);
  letter-spacing: 0.3px;
}
.ct-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}
/* RF 频段功率条（SIM/网络 上方） */
.ct-rf-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.ct-rf-bar .el-button {
  margin-left: 0;
}
.ct-rf-hint {
  margin-left: 8px;
  font-size: 12px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.55);
}
/* info-section-title 横向布局（左侧标题 + 右侧 hint，视觉压扁） */
.info-section-title-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  margin-top: 6px;
  margin-bottom: 10px;
}
.info-section-title-main {
  font-size: 13px;
  font-weight: 700;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.95));
  letter-spacing: 0.3px;
}
.info-section-title-hint {
  font-size: 11px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.5);
  letter-spacing: 0.2px;
}
/* 当前 RF 状态小标记 */
.rf-state-tag {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.3px;
  padding: 2px 10px;
  border-radius: 999px;
  margin-left: 2px;
  white-space: nowrap;
}
.rf-state-tag.on {
  color: #b9f6ca;
  background: rgba(46, 204, 113, 0.16);
  border: 1px solid rgba(46, 204, 113, 0.4);
}
.rf-state-tag.off {
  color: rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.18);
}
.rf-state-tag.warn {
  color: #ffd8a8;
  background: rgba(255, 169, 77, 0.16);
  border: 1px solid rgba(255, 169, 77, 0.4);
}
/* 漂移 / 不一致警示 */
.rf-drift-warn {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin: 8px 0 2px;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  color: #ffd8a8;
  background: rgba(255, 169, 77, 0.12);
  border: 1px solid rgba(255, 169, 77, 0.35);
}
.rf-drift-warn .warn-ico {
  flex: none;
  font-size: 13px;
  line-height: 1.4;
}
/* 状态说明（寄存器 vs 实际发射功率） */
.rf-note {
  margin-top: 8px;
  font-size: 11px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.45);
}

/* 统一操作按钮（conntrack / RF 共用），按语义配色 + 强化可点质感 */
.sys-action-btn.el-button {
  border-radius: 10px;
  font-weight: 600;
  letter-spacing: 0.3px;
  border: 1px solid transparent !important;
  color: var(--theme-text-color, #ffffff) !important;
  background: rgba(255, 255, 255, 0.10) !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.12);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: transform 0.15s ease, background 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}
.sys-action-btn .btn-ico {
  font-size: 13px;
  line-height: 1;
  opacity: 0.95;
}
.sys-action-btn.el-button:hover:not(:disabled) {
  transform: translateY(-1px);
}

/* 语义色：开启 / 生效 → 绿 */
.sys-action-btn.is-green {
  background: rgba(45, 212, 122, 0.20) !important;
  border-color: rgba(45, 212, 122, 0.60) !important;
  box-shadow: 0 0 0 1px rgba(45, 212, 122, 0.25), 0 4px 14px rgba(45, 212, 122, 0.25);
}
.sys-action-btn.is-green:hover:not(:disabled) {
  background: rgba(45, 212, 122, 0.34) !important;
  box-shadow: 0 0 0 1px rgba(45, 212, 122, 0.42), 0 6px 18px rgba(45, 212, 122, 0.35);
}
/* 语义色：移除 / 卸载 → 红 */
.sys-action-btn.is-red {
  background: rgba(255, 107, 107, 0.16) !important;
  border-color: rgba(255, 107, 107, 0.55) !important;
  box-shadow: 0 0 0 1px rgba(255, 107, 107, 0.22), 0 4px 14px rgba(255, 107, 107, 0.22);
}
.sys-action-btn.is-red:hover:not(:disabled) {
  background: rgba(255, 107, 107, 0.30) !important;
  box-shadow: 0 0 0 1px rgba(255, 107, 107, 0.42), 0 6px 18px rgba(255, 107, 107, 0.32);
}
/* 语义色：应用高功率 → 橙 */
.sys-action-btn.is-amber {
  background: rgba(255, 169, 77, 0.20) !important;
  border-color: rgba(255, 169, 77, 0.60) !important;
  box-shadow: 0 0 0 1px rgba(255, 169, 77, 0.25), 0 4px 14px rgba(255, 169, 77, 0.25);
}
.sys-action-btn.is-amber:hover:not(:disabled) {
  background: rgba(255, 169, 77, 0.34) !important;
  box-shadow: 0 0 0 1px rgba(255, 169, 77, 0.42), 0 6px 18px rgba(255, 169, 77, 0.35);
}
/* 语义色：复位 / 默认 → 蓝灰 */
.sys-action-btn.is-slate {
  background: rgba(116, 192, 252, 0.14) !important;
  border-color: rgba(116, 192, 252, 0.45) !important;
  box-shadow: 0 0 0 1px rgba(116, 192, 252, 0.20), 0 4px 14px rgba(116, 192, 252, 0.18);
}
.sys-action-btn.is-slate:hover:not(:disabled) {
  background: rgba(116, 192, 252, 0.26) !important;
  box-shadow: 0 0 0 1px rgba(116, 192, 252, 0.38), 0 6px 18px rgba(116, 192, 252, 0.28);
}

/* 禁用态：弱化到明显"不可用"，无发光、低对比，覆盖上方语义色 */
.sys-action-btn.el-button:disabled {
  opacity: 1 !important;
  cursor: not-allowed;
  background: rgba(255, 255, 255, 0.05) !important;
  border-color: rgba(255, 255, 255, 0.12) !important;
  color: rgba(255, 255, 255, 0.40) !important;
  box-shadow: none !important;
}
.sys-action-btn.el-button:disabled .btn-ico {
  opacity: 0.4;
}
.ct-main {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 16px;
  align-items: center;
  margin-bottom: 10px;
}
.ct-count {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 96px;
}
.ct-count-num {
  font-size: 30px;
  font-weight: 800;
  line-height: 1.1;
  color: var(--theme-text-color, #fff);
  font-variant-numeric: tabular-nums;
}
.ct-count-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.6));
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-top: 2px;
}
.ct-usage {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.ct-usage-bar {
  height: 10px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.15);
  overflow: hidden;
}
.ct-usage-fill {
  height: 100%;
  border-radius: 6px;
  transition: width 0.4s ease;
}
.ct-fill-safe {
  background: #68d391;
}
.ct-fill-warn {
  background: #f6ad55;
}
.ct-fill-danger {
  background: #fc8181;
}
.ct-usage-text {
  font-size: 12px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.7));
  font-variant-numeric: tabular-nums;
}
.ct-section-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.78));
  margin: 8px 0 6px;
  letter-spacing: 0.3px;
}
.ct-proto {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.ct-proto-chip {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--theme-text-color, rgba(255, 255, 255, 0.9));
}
.ct-proto-chip b {
  font-variant-numeric: tabular-nums;
  margin-left: 4px;
}
.ct-proto-empty,
.ct-top-empty {
  font-size: 12px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.5));
}
.ct-top {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.ct-top-item {
  display: grid;
  grid-template-columns: 20px 1fr auto auto;
  align-items: center;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.15);
}
.ct-top-item.is-top {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.35);
}
/* 非 LAN 行（CPE 自身上行/外网连接）：用虚线 + 淡化，区别于 LAN 设备 */
.ct-top-item.is-nonlan {
  border-style: dashed;
  background: rgba(255, 255, 255, 0.03);
  opacity: 0.78;
}
.ct-top-item.is-nonlan .ct-name {
  font-weight: 500;
  font-size: 12px;
}
.ct-top-empty-sub {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
}
.ct-rank {
  font-size: 12px;
  font-weight: 700;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.6));
  text-align: center;
}
.ct-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--theme-text-color, #fff);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ct-ip {
  font-size: 11px;
  color: var(--theme-text-color, rgba(255, 255, 255, 0.5));
  font-variant-numeric: tabular-nums;
}
.ct-cnt {
  font-size: 13px;
  font-weight: 700;
  color: #6be4ff;
  font-variant-numeric: tabular-nums;
  min-width: 48px;
  text-align: right;
}

/* 系统操作结果弹窗（装机/卸载/RF 等动作的后端细节展示） */
.action-result-summary {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  margin-bottom: 14px;
}
.action-result-summary.is-ok {
  background: rgba(45, 212, 122, 0.12);
  border: 1px solid rgba(45, 212, 122, 0.4);
  color: #b9f6ca;
}
.action-result-summary.is-fail {
  background: rgba(255, 107, 107, 0.12);
  border: 1px solid rgba(255, 107, 107, 0.4);
  color: #ffd1d1;
}
.action-result-icon {
  font-size: 18px;
  line-height: 1;
  flex: none;
}
.action-result-section {
  margin-bottom: 14px;
}
.action-result-section:last-child {
  margin-bottom: 0;
}
.action-result-section-title {
  font-size: 12px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.78);
  letter-spacing: 0.4px;
  margin-bottom: 6px;
  padding-bottom: 4px;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.18);
}
.action-result-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.action-result-item {
  display: grid;
  grid-template-columns: minmax(140px, 0.9fr) 1fr auto;
  gap: 10px;
  padding: 6px 10px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 12.5px;
  line-height: 1.45;
  align-items: baseline;
}
.action-result-item.ok {
  border-left: 3px solid rgba(45, 212, 122, 0.7);
}
.action-result-item.fail {
  border-left: 3px solid rgba(255, 107, 107, 0.7);
}
.action-result-item.warn {
  border-left: 3px solid rgba(255, 169, 77, 0.7);
}
.ar-label {
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
  word-break: break-all;
}
.ar-value {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: wrap;
  min-width: 0;
  color: rgba(255, 255, 255, 0.92);
}
.ar-pre {
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
.action-result-item code {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  background: rgba(255, 255, 255, 0.10);
  color: #ffd86b;
  font-variant-numeric: tabular-nums;
  margin-right: 4px;
  flex: none;
}
.action-result-item.ok code { color: #b9f6ca; }
.action-result-item.fail code { color: #ffbcbc; }
.ar-empty {
  color: rgba(255, 255, 255, 0.4);
  font-style: italic;
}
.action-result-raw {
  background: rgba(0, 0, 0, 0.32);
  color: #d6e1ee;
  padding: 10px 12px;
  border-radius: 8px;
  font-family: ui-monospace, "SF Mono", "JetBrains Mono", "Cascadia Code", Menlo, Consolas, monospace;
  font-size: 11.5px;
  line-height: 1.45;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 220px;
  overflow: auto;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
</style>
