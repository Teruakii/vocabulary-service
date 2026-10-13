<template>
  <li class="card word-item">
    <form v-if="editing" class="stack" @submit.prevent="save">
      <div class="edit-grid">
        <div class="field">
          <label :for="`edit-word-${word.id}`">คำศัพท์</label>
          <input :id="`edit-word-${word.id}`" v-model="draftWord" class="input" required :disabled="saving" />
        </div>
        <div class="field">
          <label :for="`edit-meaning-${word.id}`">ความหมาย</label>
          <input :id="`edit-meaning-${word.id}`" v-model="draftMeaning" class="input" required :disabled="saving" />
        </div>
      </div>
      <div class="row">
        <button type="submit" class="btn-secondary" :disabled="saving || !draftValid">
          {{ saving ? 'กำลังบันทึก…' : 'บันทึก' }}
        </button>
        <button type="button" class="btn-ghost" :disabled="saving" @click="editing = false">ยกเลิก</button>
      </div>
      <p v-if="editError" class="alert-error" role="alert">{{ editError }}</p>
    </form>

    <template v-else>
      <div class="row-between">
        <div class="word-text">
          <div class="row">
            <strong class="word">{{ word.word }}</strong>
            <span class="category-chip">{{ word.ownerUserId === null ? 'ส่วนกลาง' : 'ส่วนตัว' }}</span>
          </div>
          <span class="caption">{{ word.meaning }}</span>
        </div>
        <div class="row">
          <span v-if="added" class="badge-mastered">
            <Icon name="check" :size="16" />
            อยู่ในคำของฉันแล้ว
          </span>
          <button v-else type="button" class="btn-secondary" :disabled="adding" @click="addToMine">
            {{ adding ? 'กำลังเพิ่ม…' : 'เพิ่มเข้าคำของฉัน' }}
          </button>
          <template v-if="canEdit">
            <button type="button" class="btn-ghost" :aria-label="`แก้ไข ${word.word}`" @click="startEdit">
              <Icon name="edit" :size="20" />
            </button>
            <button
              v-if="!confirmingDelete"
              type="button"
              class="btn-ghost"
              :aria-label="`ลบ ${word.word}`"
              @click="confirmingDelete = true"
            >
              <Icon name="delete" :size="20" />
            </button>
          </template>
        </div>
      </div>

      <div v-if="confirmingDelete" class="row">
        <span class="caption">ลบ "{{ word.word }}" ออกจากคลังถาวร?</span>
        <button type="button" class="btn-danger" :disabled="deleting" @click="remove">
          {{ deleting ? 'กำลังลบ…' : 'ยืนยันลบ' }}
        </button>
        <button type="button" class="btn-ghost" :disabled="deleting" @click="cancelDelete">ยกเลิก</button>
      </div>
      <p v-if="deleteError" class="alert-error" role="alert">{{ deleteError }}</p>
    </template>
  </li>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/Icon.vue'
import { userWords, words, type Word } from '@/api/client'
import { useToast } from '@/composables/useToast'
import { isStatus, messageOf } from '@/api/client'

const props = defineProps<{ word: Word; canEdit: boolean }>()
const emit = defineEmits<{ updated: [word: Word]; removed: [id: number] }>()

const { push } = useToast()

const adding = ref(false)
// The API has no "is this in my list" lookup, so a 409 on add is how we learn it already is.
const added = ref(false)

async function addToMine() {
  adding.value = true
  try {
    await userWords.add({ wordId: props.word.id })
    added.value = true
  } catch (error) {
    if (isStatus(error, 409)) added.value = true
    else push(messageOf(error))
  } finally {
    adding.value = false
  }
}

const editing = ref(false)
const draftWord = ref('')
const draftMeaning = ref('')
const saving = ref(false)
const editError = ref('')
const draftValid = computed(() => draftWord.value.trim() !== '' && draftMeaning.value.trim() !== '')

function startEdit() {
  draftWord.value = props.word.word
  draftMeaning.value = props.word.meaning
  editError.value = ''
  confirmingDelete.value = false
  editing.value = true
}

async function save() {
  if (!draftValid.value) return
  saving.value = true
  editError.value = ''
  try {
    const updated = await words.update(props.word.id, {
      word: draftWord.value.trim(),
      meaning: draftMeaning.value.trim(),
    })
    emit('updated', updated)
    editing.value = false
  } catch (error) {
    editError.value = messageOf(error)
  } finally {
    saving.value = false
  }
}

const confirmingDelete = ref(false)
const deleting = ref(false)
const deleteError = ref('')

function cancelDelete() {
  confirmingDelete.value = false
  deleteError.value = ''
}

async function remove() {
  deleting.value = true
  deleteError.value = ''
  try {
    await words.remove(props.word.id)
    emit('removed', props.word.id)
  } catch (error) {
    deleteError.value = messageOf(error)
    deleting.value = false
  }
}
</script>

<style scoped>
.word-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
}

.word-text {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  min-width: 0;
}

.word {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
  overflow-wrap: anywhere;
}

.edit-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 2fr);
  gap: var(--space-md);
}

@media (max-width: 639px) {
  .edit-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
