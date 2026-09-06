<script setup lang="ts">
type Device = { id: number; name: string; mac: string }
const devices = ref<Device[]>([]), loading = ref(true), error = ref(''), showForm = ref(false)
const form = reactive({ name: '', mac: '' })
async function load() { loading.value = true; try { devices.value = await $fetch<Device[]>('/api/devices') } catch { error.value = 'No se pudo conectar con la API.' } finally { loading.value = false } }
async function wake(id: number) { await $fetch(`/api/devices/${id}/wake`, { method: 'POST' }); await load() }
async function add() {
    error.value = ''
    try {
        await $fetch('/api/devices', { method: 'POST', body: form })
        Object.assign(form, { name: '', mac: '' })
        showForm.value = false
        await load()
    } catch (cause: any) {
        console.error('Error guardando dispositivo', cause)
        error.value = cause?.data?.error ?? cause?.statusMessage ?? cause?.message ?? 'No se pudo guardar el dispositivo.'
    }
}
async function remove(id: number) { if (confirm('¿Eliminar este dispositivo?')) { await $fetch(`/api/devices/${id}`, { method: 'DELETE' }); await load() } }
onMounted(load)
</script>
<template>
    <main class="page">
        <header>
            <div>
                <p class="eyebrow">WAKE CONTROL</p>
                <h1>Mis dispositivos</h1>
                <p class="muted">Enciende tus equipos con un toque.</p>
            </div><button class="primary" @click="showForm = !showForm">＋ Añadir</button>
        </header>
        <form v-if="showForm" class="form card" @submit.prevent="add"><input v-model="form.name" placeholder="Nombre"
                required><input v-model="form.mac" placeholder="AA:BB:CC:DD:EE:FF"
                    pattern="^([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}$"
                title="Formato esperado: AA:BB:CC:DD:EE:FF" required><button class="primary">Guardar</button></form>
        <p v-if="error" class="error">{{ error }}</p>
        <section class="grid">
            <article v-for="device in devices" :key="device.id" class="card device">
                <div class="status"><span />Dispositivo registrado</div>
                <h2>{{ device.name }}</h2>
                <p class="muted">{{ device.mac }}</p>
                <div class="actions"><button class="wake" @click="wake(device.id)">⚡ Despertar</button><button
                        class="delete" @click="remove(device.id)">×</button></div>
            </article>
            <div v-if="!loading && !devices.length" class="empty card">Añade tu primer dispositivo.</div>
        </section>
    </main>
</template>
<style>
:root {
    font-family: system-ui, sans-serif;
    color: #e8edf5;
    background: #0b1020
}

body {
    margin: 0
}

.page {
    max-width: 1100px;
    margin: auto;
    padding: 48px 24px
}

header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 32px
}

.eyebrow {
    color: #8b9cff;
    letter-spacing: .18em;
    font-size: .75rem;
    font-weight: 700
}

h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    margin: .2rem 0
}

.muted {
    color: #8f9bb1
}

.grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 20px
}

.card {
    background: #141b2d;
    border: 1px solid #26314a;
    border-radius: 20px;
    padding: 24px
}

.device {
    min-height: 190px;
    display: flex;
    flex-direction: column
}

.device h2 {
    font-size: 1.5rem;
    margin: 28px 0 8px
}

.status {
    color: #ff8e8e;
    display: flex;
    gap: 8px;
    align-items: center
}

.status span {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: currentColor
}

.actions {
    margin-top: auto;
    display: flex;
    gap: 10px
}

button {
    border: 0;
    border-radius: 12px;
    padding: 13px 18px;
    font-weight: 700;
    cursor: pointer
}

.primary {
    background: #7185ff;
    color: white
}

.wake {
    background: #263b67;
    color: #b9c9ff;
    flex: 1
}

.delete {
    background: transparent;
    color: #a6aec0;
    font-size: 1.4rem
}

.form {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 24px
}

input {
    background: #0e1526;
    border: 1px solid #34405d;
    border-radius: 10px;
    padding: 13px;
    color: white;
    min-width: 160px
}

.error {
    color: #ff8e8e
}

.empty {
    grid-column: 1/-1;
    text-align: center;
    padding: 60px
}
</style>
