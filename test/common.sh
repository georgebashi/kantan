
function start_kantan {
	test_expect_success 'start kantan' '../../kantan &'
	KANTAN_PID=$!
}

function stop_kantan {
	test_expect_success 'stop kantan' 'kill $KANTAN_PID'
}
