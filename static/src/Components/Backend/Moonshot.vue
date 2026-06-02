<script setup>
import { ref } from 'vue';
import {Setting} from "@/js/util.js";

const settings = ref()

async function getSetting() {
  settings.value = await Setting("fetchSettings", "moonshot")
}

async function sendSetting() {
  await Setting("editSettings", "moonshot", settings.value)
}
</script>

<template>
  <mdui-collapse>
    <mdui-collapse-item rounded>
      <mdui-list-item slot="header" icon="settings_applications" rounded @click="getSetting()">
        Moonshot AI (Kimi)
        <mdui-icon slot="end-icon" name="keyboard_arrow_down"></mdui-icon>
      </mdui-list-item>
      <mdui-list-item nonclickable>
        <mdui-card variant="outlined">
          <p>这里设置Moonshot AI（Kimi）的后端API，可用于文字生成，总结等</p>
          <mdui-text-field variant="outlined" label="API Key" type="password" toggle-password
                            :value="settings?.api_key || ''"
                            @change="settings.api_key = $event.target.value"></mdui-text-field>
          <mdui-text-field variant="outlined" label="API地址"
                            :value="settings?.endpoint || 'https://api.moonshot.cn/v1/chat/completions'"
                            @change="settings.endpoint = $event.target.value"></mdui-text-field>
          <p>默认文字生成模型</p>
          <mdui-radio-group :value="settings?.text_model || 'kimi-for-coding'" style="margin-top: 0"
                            @change="settings.text_model=$event.target.value">
            <mdui-radio value="kimi-for-coding">Kimi for Coding（编程专用）</mdui-radio>
            <mdui-radio value="moonshot-v1-8k">Moonshot V1 8K</mdui-radio>
            <mdui-radio value="moonshot-v1-32k">Moonshot V1 32K</mdui-radio>
            <mdui-radio value="moonshot-v1-128k">Moonshot V1 128K</mdui-radio>
          </mdui-radio-group>
          <p>默认总结模型</p>
          <mdui-radio-group :value="settings?.summary_model || 'kimi-for-coding'" style="margin-top: 0"
                            @change="settings.summary_model=$event.target.value">
            <mdui-radio value="kimi-for-coding">Kimi for Coding（编程专用）</mdui-radio>
            <mdui-radio value="moonshot-v1-8k">Moonshot V1 8K</mdui-radio>
            <mdui-radio value="moonshot-v1-32k">Moonshot V1 32K</mdui-radio>
            <mdui-radio value="moonshot-v1-128k">Moonshot V1 128K</mdui-radio>
          </mdui-radio-group>
          <p>最大Token数</p>
          <mdui-text-field variant="outlined" label="Max Tokens"
                            :value="settings?.max_tokens || 8192"
                            @change="settings.max_tokens = parseInt($event.target.value) || 8192"
                            type="number"></mdui-text-field>
          <p>温度参数 (0-1，推荐编程用0.3)</p>
          <mdui-slider :value="settings?.temperature || 0.3" min="0" max="1" step="0.1"
                        @change="settings.temperature = $event.target.value"></mdui-slider>
          <p>Top-P (0-1)</p>
          <mdui-slider :value="settings?.top_p || 0.95" min="0" max="1" step="0.05"
                        @change="settings.top_p = $event.target.value"></mdui-slider>
          <p>启用状态</p>
          <mdui-switch :checked="settings?.enabled !== false"
                        @change="settings.enabled = $event.target.checked"></mdui-switch>
          <mdui-button @click="sendSetting()">确认</mdui-button>
        </mdui-card>
      </mdui-list-item>
    </mdui-collapse-item>
  </mdui-collapse>
</template>

<style scoped>

</style>
