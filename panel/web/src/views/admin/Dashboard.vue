<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>数据看板</h2>
        <div class="sub">平台整体运行态势一览</div>
      </div>
      <n-button size="small" @click="loadData">
        <template #icon><svg-icon name="refresh" :size="14" /></template>
        刷新
      </n-button>
    </div>

    <div class="stat-row">
      <div class="glass-card mini-stat">
        <div class="stat-icon cyan"><svg-icon name="users" :size="20" /></div>
        <div>
          <div class="stat-label">总用户数</div>
          <div class="stat-value">{{ stats.total_users ?? 0 }}</div>
        </div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-icon green"><svg-icon name="plus" :size="20" /></div>
        <div>
          <div class="stat-label">今日新增</div>
          <div class="stat-value">{{ stats.new_users ?? 0 }}</div>
        </div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-icon orange"><svg-icon name="node" :size="20" /></div>
        <div>
          <div class="stat-label">在线节点</div>
          <div class="stat-value">{{ stats.online_nodes ?? 0 }}</div>
        </div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-icon warm"><svg-icon name="tunnel" :size="20" /></div>
        <div>
          <div class="stat-label">在线隧道</div>
          <div class="stat-value">{{ stats.online_tunnels ?? 0 }}</div>
        </div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-icon orange"><svg-icon name="points" :size="20" /></div>
        <div>
          <div class="stat-label">积分发放</div>
          <div class="stat-value">{{ stats.points_issued ?? 0 }}</div>
        </div>
      </div>
    </div>

    <div class="glass-card chart-card">
      <div class="chart-title">近 7 天流量趋势</div>
      <div ref="chartEl" class="chart"></div>
    </div>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { NButton } from 'naive-ui'
import * as echarts from 'echarts'
import SvgIcon from '@/components/SvgIcon.vue'
import { adminGetDashboard } from '@/api'
import { formatBytes } from '@/utils/format'

const chartEl = ref(null)
let chart = null
const stats = reactive({})

function parseTraffic(data) {
  let dates = []
  let inData = []
  let outData = []
  if (Array.isArray(data)) {
    data.forEach((item) => {
      dates.push(item.date || item.day || '')
      inData.push(Number(item.in_bytes ?? item.in ?? 0))
      outData.push(Number(item.out_bytes ?? item.out ?? 0))
    })
  } else if (data && Array.isArray(data.dates)) {
    dates = data.dates
    inData = (data.in || data.in_bytes || []).map(Number)
    outData = (data.out || data.out_bytes || []).map(Number)
  }
  return { dates, inData, outData }
}

function renderChart(data) {
  if (!chartEl.value) return
  if (!chart) chart = echarts.init(chartEl.value)
  const { dates, inData, outData } = parseTraffic(data)
  chart.setOption(
    {
      color: ['#0ea5e9', '#f97316'],
      tooltip: { trigger: 'axis', valueFormatter: (v) => formatBytes(v) },
      legend: { data: ['上行', '下行'], bottom: 0, icon: 'roundRect', itemWidth: 14, itemHeight: 4 },
      grid: { left: 60, right: 20, top: 30, bottom: 40 },
      xAxis: {
        type: 'category',
        data: dates,
        boundaryGap: false,
        axisLine: { lineStyle: { color: '#cbd5e1' } },
        axisLabel: { color: '#64748b' }
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#64748b' },
        splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.2)' } }
      },
      series: [
        {
          name: '上行',
          type: 'line',
          smooth: true,
          symbolSize: 6,
          data: inData,
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(14, 165, 233, 0.28)' },
              { offset: 1, color: 'rgba(14, 165, 233, 0.02)' }
            ])
          }
        },
        {
          name: '下行',
          type: 'line',
          smooth: true,
          symbolSize: 6,
          data: outData,
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(249, 115, 22, 0.24)' },
              { offset: 1, color: 'rgba(249, 115, 22, 0.02)' }
            ])
          }
        }
      ]
    },
    true
  )
}

function handleResize() {
  chart?.resize()
}

async function loadData() {
  try {
    const data = (await adminGetDashboard()) || {}
    if (Array.isArray(data.traffic_series) || data?.dates) {
      renderChart(data.traffic_series || data)
    }
    const s = data.stats || data
    if (s && typeof s === 'object') {
      Object.assign(stats, s)
    }
  } catch (e) {
    // 错误已由拦截器提示
  }
}

onMounted(() => {
  loadData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
  chart = null
})
</script>

<style scoped>
.stat-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.mini-stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 16px;
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.cyan {
  background: rgba(14, 165, 233, 0.14);
  color: #0ea5e9;
}

.stat-icon.green {
  background: rgba(16, 185, 129, 0.14);
  color: #10b981;
}

.stat-icon.orange {
  background: rgba(249, 115, 22, 0.14);
  color: #f97316;
}

.stat-icon.warm {
  background: rgba(245, 158, 11, 0.16);
  color: #f59e0b;
}

.chart-card {
  padding: 20px;
}

.chart-title {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 12px;
}

.chart {
  width: 100%;
  height: 380px;
}

@media (max-width: 1100px) {
  .stat-row {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
