/* placeholder file for JavaScript */
const confirm_task_delete = (id, title) => {
    if(window.confirm(`Task ${id} ${title} を削除します．よろしいですか？`)) {
        // document.getElementById('task_delete').submit();
        fetch(`/task/delete/${id}`, {
            method: 'DELETE',
        }).then((res) => {
            if(res.status === 200) {
                window.location.href = '/list';
            } else {
                alert('削除に失敗しました．');
            }
        });
    }
}

const confirm_user_delete = (id, name) => {
    if(window.confirm(`${name}を退会させます．よろしいですか？`)) {
        // document.getElementById('user_delete').submit();
        fetch(`/user/delete`, {
            method: 'DELETE',
        }).then((res) => {
            if(res.status === 200) {
                window.location.href = '/';
            } else {
                alert('退会に失敗しました．');
            }
        });
    }
}
 
const confirm_task_update = (id, title) => {
    if(window.confirm(`Task ${id} ${title} を更新します．よろしいですか？`)) {
        document.getElementById('task_update').submit();
    }
}
