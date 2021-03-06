# -*- coding: utf-8 -*-
"""TTS eb3t l yoush as python file.ipynb

Automatically generated by Colaboratory.

Original file is located at
    https://colab.research.google.com/drive/1m-h4rsKPxgdXQOjsRk0JQAZKYO4zixqE
"""

import os
# !git clone https://github.com/TensorSpeech/TensorFlowTTS
# os.chdir("TensorFlowTTS")
# !pip install  .
# os.chdir("..")
import sys
import gdown
# sys.path.append("TensorFlowTTS/")

print("Downloading Tacotron2 model...")
#gdown --id {"12jvEO1VqFo1ocrgY9GUHF_kVcLn3QaGW"} -O tacotron2-120k.h5
#gdown --id {"1OI86hkN1YCpHBsIKnkELNbSho5Pj-pPY"} -O tacotron2_config.yml

print("Downloading Multi-band MelGAN model...")
#gdown --id {"1kChFaLI7slrTtuk3pvcOiJwJDCygsw9C"} -O mb.melgan-940k.h5
#gdown --id {"1YC_kZpuRZGQ-JfMKj1LC0YRyKXsgLTJL"} -O mb.melgan_config.yml

import tensorflow as tf

import yaml
import numpy as np
from scipy.io.wavfile import write
#import matplotlib.pyplot as plt

#import IPython.display as ipd


from tensorflow_tts.inference.auto_model import TFAutoModel
from tensorflow_tts.inference.auto_config import AutoConfig
from tensorflow_tts.inference.auto_processor import AutoProcessor



tacotron2_config = AutoConfig.from_pretrained('TensorFlowTTS/examples/tacotron2/conf/tacotron2.v1.yaml')
tacotron2 = TFAutoModel.from_pretrained(
    config=tacotron2_config,
    pretrained_path="tacotron2-120k.h5",
    name="tacotron2"
)

mb_melgan_config = AutoConfig.from_pretrained('TensorFlowTTS/examples/multiband_melgan/conf/multiband_melgan.v1.yaml')
mb_melgan = TFAutoModel.from_pretrained(
    config=mb_melgan_config,
    pretrained_path="mb.melgan-940k.h5",
    name="mb_melgan"
)

print("Downloading ljspeech_mapper.json ...")
# !gdown --id {"1YBaDdMlhTXxsKrH7mZwDu-2aODq5fr5e"} -O ljspeech_mapper.json

processor = AutoProcessor.from_pretrained(pretrained_path="./ljspeech_mapper.json")

def do_synthesis(input_text, text2mel_model, vocoder_model, text2mel_name, vocoder_name):
  input_ids = processor.text_to_sequence(input_text)

  # text2mel part
  if text2mel_name == "TACOTRON":
    _, mel_outputs, stop_token_prediction, alignment_history = text2mel_model.inference(
        tf.expand_dims(tf.convert_to_tensor(input_ids, dtype=tf.int32), 0),
        tf.convert_to_tensor([len(input_ids)], tf.int32),
        tf.convert_to_tensor([0], dtype=tf.int32)
    )
  elif text2mel_name == "FASTSPEECH":
    mel_before, mel_outputs, duration_outputs = text2mel_model.inference(
        input_ids=tf.expand_dims(tf.convert_to_tensor(input_ids, dtype=tf.int32), 0),
        speaker_ids=tf.convert_to_tensor([0], dtype=tf.int32),
        speed_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
    )
  elif text2mel_name == "FASTSPEECH2":
    mel_before, mel_outputs, duration_outputs, _, _ = text2mel_model.inference(
        tf.expand_dims(tf.convert_to_tensor(input_ids, dtype=tf.int32), 0),
        speaker_ids=tf.convert_to_tensor([0], dtype=tf.int32),
        speed_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
        f0_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
        energy_ratios=tf.convert_to_tensor([1.0], dtype=tf.float32),
    )
  else:
    raise ValueError("Only TACOTRON, FASTSPEECH, FASTSPEECH2 are supported on text2mel_name")

  # vocoder part
  if vocoder_name == "MELGAN" or vocoder_name == "MELGAN-STFT":
    audio = vocoder_model(mel_outputs)[0, :, 0]
  elif vocoder_name == "MB-MELGAN":
    audio = vocoder_model(mel_outputs)[0, :, 0]
  else:
    raise ValueError("Only MELGAN, MELGAN-STFT and MB_MELGAN are supported on vocoder_name")

  if text2mel_name == "TACOTRON":
    return mel_outputs.numpy(), alignment_history.numpy(), audio.numpy()
  else:
    return mel_outputs.numpy(), audio.numpy()

# def visualize_attention(alignment_history):
#   import matplotlib.pyplot as plt

#   fig = plt.figure(figsize=(8, 6))
#   ax = fig.add_subplot(111)
#   ax.set_title(f'Alignment steps')
#   im = ax.imshow(
#       alignment_history,
#       aspect='auto',
#       origin='lower',
#       interpolation='none')
#   fig.colorbar(im, ax=ax)
#   xlabel = 'Decoder timestep'
#   plt.xlabel(xlabel)
#   plt.ylabel('Encoder timestep')
#   plt.tight_layout()
#   plt.show()
#   plt.close()

# def visualize_mel_spectrogram(mels):
#   mels = tf.reshape(mels, [-1, 80]).numpy()
#   fig = plt.figure(figsize=(10, 8))
#   ax1 = fig.add_subplot(311)
#   ax1.set_title(f'Predicted Mel-after-Spectrogram')
#   im = ax1.imshow(np.rot90(mels), aspect='auto', interpolation='none')
#   fig.colorbar(mappable=im, shrink=0.65, orientation='horizontal', ax=ax1)
#   plt.show()
#   plt.close()


input_text = sys.argv[1]

# setup window for tacotron2 if you want to try
tacotron2.setup_window(win_front=10, win_back=10)



mels, alignment_history, audios = do_synthesis(input_text, tacotron2, mb_melgan, "TACOTRON", "MB-MELGAN")
# #visualize_attention(alignment_history[0])
# #visualize_mel_spectrogram(mels[0])

print("SYS arg 2: ", sys.argv[2])
print("BEFORE WRITE")

write("posts_audio/"+ sys.argv[2], 21550,audios )

print(sys.argv[1], sys.argv[2])