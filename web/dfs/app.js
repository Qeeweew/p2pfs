async function updateList() {
    const res = await fetch('/api/files');
    const files = await res.json();
    const ul = document.getElementById('fileList');
    ul.innerHTML = '';
    files.forEach(f => {
        const li = document.createElement('li');
        const a = document.createElement('a');
        a.href = `/api/download?file=${encodeURIComponent(f)}`;
        a.textContent = f;
        li.appendChild(a);
        ul.appendChild(li);
    });
}

async function upload() {
    const input = document.getElementById('fileInput');
    if (!input.files.length) return;
    const file = input.files[0];
    const form = new FormData();
    form.append('file', file);
    await fetch('/api/upload', { method: 'POST', body: form });
    updateList();
}

window.onload = updateList;
