<template>
  <div class="page-container">
    <div class="page-title">
      <div>
        <h2>流量统计</h2>
        <div class="sub">近 7 天隧道流量趋势与今日概览</div>
      </div>
      <n-button size="small" @click="loadData">
        <template #icon><svg-icon name="refresh" :size="14" /></template>
        刷新数据
      </n-button>
    </div>

    <div class="stat-row">
      <div class="glass-card mini-stat">
        <div class="stat-label">今日上行</div>
        <div class="stat-value">{{ formatBytes(today.in) }}</div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-label">今日下行</div>
        <div class="stat-value">{{ formatBytes(today.out) }}</div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-label">今日合计</div>
        <div class="stat-value">{{ formatBytes(today.total) }}</div>
      </div>
      <div class="glass-card mini-stat">
        <div class="stat-label">7 天合计</div>
        <div class="stat-value">{{ formatBytes(total7d) }}</div>
      </div>
    </div>

    <div class="glass-card chart-card">
      <div ref="chartEl" class="chart"></div>
    </div>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { NButton } from 'naive-ui'
import * as echarts from 'echarts'
import SvgIcon from '@/components/SvgIcon.vue'
import { getTrafficStats, getStatsOverview } from '@/api'
import { formatBytes } from '@/utils/format'

const chartEl = ref(null)
let chart = null
const today = reactive({ in: 0, out: 0, total: 0 })
const total7d = ref(0)

function parseTraffic(data) {
  let dates = []
  let inData = []
  let outData = []
  if (Array.isArray(data)) {
    data.forEach((item) => {
      dates.push(item.date || item.day || item.ds || '')
      inData.push(Number(item.in_bytes ?? item.in ?? item.up ?? 0))
      outData.push(Number(item.out_bytes ?? item.out ?? item.down ?? 0))
    })
  } else if (data && Array.isArray(data.dates)) {
    dates = data.dates
    inData = (data.in || data.in_bytes || []).map(Number)
    outData = (data.out || data.out_bytes || []).map(Number)
  } else if (data && Array.isArray(data.traffic_series)) {
    data.traffic_series.forEach((item) => {
      dates.push(item.date || item.day || '')
      inData.push(Number(item.in_bytes ?? item.in ?? 0))
      outData.push(Number(item.out_bytes ?? item.out ?? 0))
    })
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
      tooltip: {
        trigger: 'axis',
        valueFormatter: (v) => formatBytes(v)
      },
      legend: {
        data: ['上行', '下行'],
        bottom: 0,
        icon: 'roundRect',
        itemWidth: 14,
        itemHeight: 4
      },
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
          symbol: 'circle',
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
          symbol: 'circle',
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
    const [traffic, overview] = await Promise.allSettled([
      getTrafficStats({ days: 7 }),
      getStatsOverview()
    ])
    if (traffic.status === 'fulfilled') renderChart(traffic.value)
    if (overview.status === 'fulfilled' && overview.value) {
      const o = overview.value
      today.in = Number(o.today_in ?? o.today_up ?? o.in ?? 0)
      today.out = Number(o.today_out ?? o.today_down ?? o.out ?? 0)
      today.total = Number(o.today_total ?? o.total ?? today.in + today.out)
    }
    if (traffic.status === 'fulfilled') {
      const { inData, outData } = parseTraffic(traffic.value)
      const sum = (arr) => arr.reduce((a, b) => a + (Number(b) || 0), 0)
      const total = sum(inData) + sum(outData)
      total7d.value = total || today.total
      if (!overview.value || overview.status === 'rejected') {
        if (inData.length) {
          today.in = Number(inData[inData.length - 1] || 0)
          today.out = Number(outData[outData.length - 1] || 0)
          today.total = today.in + today.out
        }
      }
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
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.mini-stat {
  padding: 18px 20px;
}

.chart-card {
  padding: 20px;
}

.chart {
  width: 100%;
  height: 380px;
}

@media (max-width: 1000px) {
  .stat-row {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
