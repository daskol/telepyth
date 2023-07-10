from telepyth.huggingface import format_metrics


def test_format_metrics():
    message = format_metrics({
        'epoch': 10.0,
        'train_loss': 0.6120437137211595,
        'train_runtime': 563.276,
        'train_samples_per_second': 151.808,
        'train_steps_per_second': 9.498,
    })
    assert message != ''
