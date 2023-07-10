"""Submodule huggingface defines a HuggingFace's transformers trainer callback
to report metrics directly to Telegram.

In order to use this callback one can just directly import
:py:`TelePythCallback` from this submodule or from py:`telepyth` module if it
was imported after :py:`transformers`.

.. code-block:: python

   import transformers
   from telepyth import TelePythCallback
"""

from operator import itemgetter
from typing import Dict, Optional
from datetime import datetime, timedelta

from transformers.integrations import INTEGRATION_TO_CALLBACK
from transformers.trainer_callback import (TrainerCallback, TrainerControl,
                                           TrainerState)
from transformers.training_args import TrainingArguments

from telepyth.client import TelePythClient

__all__ = ('TelePythCallback',)


def format_metrics(logs: Dict[str, float], label: Optional[str] = None) -> str:
    width = 5
    telemetry = {}
    for key, val in logs.items():
        if key == 'epoch':
            key = f'common/{key}'
        elif key.startswith('train_'):
            key = f'train/{key[6:]}'
        elif key.startswith('eval_'):
            key = f'eval/{key[5:]}'
        else:
            key = f'train/{key}'
        width = max(width, len(key))
        telemetry[key] = val

    lines = []
    if label is not None:
        lines.append(f'**Label:** `{label}`\n')
    lines.append('```')
    lines.append(f'{"Metric":{width}s} {"Value":13s}')
    lines.append('-' * len(lines[-1]))
    for key, val in sorted(telemetry.items(), key=itemgetter(0)):
        lines.append(f'{key:{width}s} {val: e}')
    lines.append('```')
    return '\n'.join(lines)


class TelePythCallback(TrainerCallback):
    """A HuggingFace's transformers trainer callback for reporting to Telegram
    messanger.

    At the moment the callback is quite simple and has two policies: each and
    last. With the first one the callback sends everything it gets via `on_log`
    method. Alternatively, the second policy can be used in order to send a
    single notification at the end of training. In this case all metrics are
    accumulated in a :py:`dict` during training and are sent as a single batch.

    Args:
      label: identifier of particular run.
      policy: when and how send notifications to Telegram.
      telepyth: option to override default telepyth client.

    There are several usage scenario. The first one is via setting reporting
    method in :py:`TrainingArguments`.

    .. code-block:: python

       args = TrainingArguments(output_dir, report_to=['telepyth'])

    For more fine-grained control on how and when callback should report
    metrics, one can set up callback directly in a list of callbacks.

    .. code-block:: python

       callback = TelePythCallback(label='finetune', policy='last')
       trainer = Trainer(callbacks=[callback])
       trainer.train()
    """

    DEFAULT_LABEL = '<unknown>'

    def __init__(self, label: Optional[str] = None, policy: str = 'last',
                 telepyth: Optional[TelePythClient] = None):
        super().__init__()
        if policy not in ('each', 'last'):
            raise ValueError(f'Unknown policy {policy}.')
        self.label = label or TelePythCallback.DEFAULT_LABEL
        self.policy = policy
        self.telepyth = telepyth or TelePythClient()
        self.metrics = {}

    def on_train_begin(self, args: TrainingArguments, state: TrainerState,
                       control: TrainerControl, **kwargs):
        self.train_begun_at = datetime.now()

    def on_train_end(self, args: TrainingArguments, state: TrainerState,
                     control: TrainerControl, **kwargs):
        self.train_ended_at = datetime.now()
        dur = self.train_ended_at - self.train_begun_at
        dur = timedelta(seconds=int(dur.total_seconds()))
        elapsed = f'Run `{self.label}` is completed in {dur}.'

        if self.policy == 'each':
            message = elapsed
        else:
            message = format_metrics(self.metrics)
            message = '\n'.join([message, elapsed])

        self.telepyth.send_text(message)

    def on_evaluate(self, args: TrainingArguments, state: TrainerState,
                    control: TrainerControl, metrics: Dict[str, float] = {},
                    **kwargs):
        pass

    def on_log(self, args: TrainingArguments, state: TrainerState,
               control: TrainerControl, logs: Dict[str, float] = {}, **kwargs):
        if self.policy == 'each':
            message = format_metrics(logs, self.label)
            self.telepyth.send_text(message)
        elif self.policy == 'last':
            self.metrics.update(logs)


# Register telepyth callback in an index of known intergrations.
INTEGRATION_TO_CALLBACK['telepyth'] = TelePythCallback
